package main

import (
	"encoding/gob"
	"log"
	"os"
	"strconv"
	"time"

	gameengine "checkerbots/apps/server/game-engine"
)

// appState is the authoritative server state, owned exclusively by the state manager goroutine.
type appState struct {
	Version     int64
	Game        gameengine.Game
	Robots      map[RobotID]robotInfo
	Assignments pieceAssignmentState
	Planner     plannerState
	UpdatedAt   time.Time
}

type RobotID string

// robotInfo holds the live state of a physical robot reported by external systems.
type robotInfo struct {
	Pose      Pose
	UpdatedAt time.Time
}

type pieceAssignmentState struct {
	robotIDByPieceID map[gameengine.PieceID]RobotID
	pieceIDByRobotID map[RobotID]gameengine.PieceID
}

func (s *pieceAssignmentState) Assign(robotID RobotID, pieceID gameengine.PieceID) {
	s.robotIDByPieceID[pieceID] = robotID
	s.pieceIDByRobotID[robotID] = pieceID
}

func (s *pieceAssignmentState) Unassign(robotID RobotID) {
	if pieceID, exists := s.pieceIDByRobotID[robotID]; exists {
		delete(s.robotIDByPieceID, pieceID)
		delete(s.pieceIDByRobotID, robotID)
	}
}

func (s *pieceAssignmentState) GetRobotIDByPieceID(pieceID gameengine.PieceID) (robotID RobotID, ok bool) {
	robotID, ok = s.robotIDByPieceID[pieceID]
	return
}

func (s *pieceAssignmentState) GetPieceIDByRobotID(robotID RobotID) (pieceID gameengine.PieceID, ok bool) {
	pieceID, ok = s.pieceIDByRobotID[robotID]
	return
}

type plannerState struct {
	Goals map[RobotID]robotGoal
}

type robotGoal struct {
	BoardPos   gameengine.Position
	IsKing     bool
	IsCaptured bool
	// TOOD: Track updated time?
}

// storedAppState is the disk-serializable form of appState.
type storedAppState struct {
	Version     int64
	Game        gameengine.StoredGame
	Robots      map[RobotID]robotInfo
	Assignments pieceAssignmentState
	Planner     plannerState
	UpdatedAt   time.Time
}

func saveState(filePath string, state appState) error {
	if filePath == "" {
		return nil
	}
	log.Printf("saving state to: %s", filePath)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	return gob.NewEncoder(file).Encode(storedAppState{
		Version:     state.Version,
		Game:        gameengine.GameToStored(state.Game),
		Robots:      state.Robots,
		Assignments: state.Assignments,
		Planner:     state.Planner,
		UpdatedAt:   state.UpdatedAt,
	})
}

func loadState(filePath string) (appState, error) {
	log.Printf("loading state from: %s", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return appState{}, err
	}
	defer file.Close()
	var stored storedAppState
	if err := gob.NewDecoder(file).Decode(&stored); err != nil {
		return appState{}, err
	}
	return appState{
		Version:     stored.Version,
		Game:        gameengine.GameFromStored(stored.Game),
		Robots:      stored.Robots,
		Assignments: stored.Assignments,
		Planner:     stored.Planner,
		UpdatedAt:   stored.UpdatedAt,
	}, nil
}

// boardCellSizeMM is the physical size of one board square in millimetres.
const boardCellSizeMM = 450.0

// positionToMM converts a board (row, col) position to mm coordinates relative
// to the centre of the board. X increases toward higher columns, Y increases
// toward higher rows.
func positionToMM(pos gameengine.Position, boardSize int) (xMM, yMM float64) {
	center := float64(boardSize-1) / 2.0
	xMM = (float64(pos.Col) - center) * boardCellSizeMM
	yMM = (float64(pos.Row) - center) * boardCellSizeMM
	return
}

// syncRobotGoals ensures state.Planner matches the current piece positions.
// It returns all RobotIDs whose goals were created or updated.
func syncRobotGoals(state *appState) []RobotID {
	var changed []RobotID

	for _, piece := range state.Game.Pieces {
		robotID, assigned := state.Assignments.GetRobotIDByPieceID(piece.ID)
		if !assigned {
			continue
		}

		newGoal := robotGoal{
			BoardPos:   piece.Position,
			IsKing:     piece.Kind == gameengine.PieceKindKing,
			IsCaptured: piece.Captured,
		}

		// TODO: Ensure that every robot always has a goal?
		if currentGoal, exists := state.Planner.Goals[robotID]; !exists || currentGoal != newGoal {
			state.Planner.Goals[robotID] = newGoal
			xMM, yMM := positionToMM(piece.Position, state.Game.BoardSize)
			state.Robots[robotID] = robotInfo{
				Pose:      Pose{XMM: xMM, YMM: yMM},
				UpdatedAt: time.Now(),
			}
			changed = append(changed, robotID)
		}
	}

	return changed
}

// --- State manager ---

// Commands sent to the state manager goroutine.

type newGameCmd struct {
	reply chan struct{}
}

type applyMoveCmd struct {
	move  gameengine.Move
	reply chan *gameengine.ApplyMoveError
}

type getSnapshotCmd struct {
	reply chan AppSnapshot
}

type saveCmd struct {
	filePath string
	reply    chan error
}

// stateManager serializes all reads and writes to appState.
// HTTP handlers interact with state exclusively via this manager's methods.
type stateManager struct {
	cmds chan any
}

func newStateManager() stateManager {
	return stateManager{cmds: make(chan any)}
}

// run is the state manager's main loop and must be called in its own goroutine.
// publish is called synchronously after each successful state change.
func (m *stateManager) run(initial appState, broadcastChan chan<- sseEvent) {
	state := initial
	for cmd := range m.cmds {
		switch c := cmd.(type) {
		case newGameCmd:
			state.Game = gameengine.NewGame8x8()
			state.Robots = map[RobotID]robotInfo{}
			state.Assignments = pieceAssignmentState{
				robotIDByPieceID: make(map[gameengine.PieceID]RobotID),
				pieceIDByRobotID: make(map[RobotID]gameengine.PieceID),
			}
			for i, piece := range state.Game.Pieces {
				robotID := RobotID("r" + strconv.Itoa(i))
				state.Robots[robotID] = robotInfo{
					// TODO: Figure out initial positions
					Pose:      Pose{},
					UpdatedAt: time.Now(),
				}
				state.Assignments.Assign(robotID, piece.ID)
			}
			state.Planner = plannerState{
				Goals: make(map[RobotID]robotGoal),
			}

			state.Version++
			state.UpdatedAt = time.Now()
			// Send a full snapshot since (right now) all robots move
			// If robot.updated events are sent, the channel buffer will overflow since ~25 events are dumped into it
			syncRobotGoals(&state)
			broadcastChan <- sseEvent{name: "state.snapshot", data: buildAppSnapshot(state)}
			c.reply <- struct{}{}

		case applyMoveCmd:
			if err := gameengine.ApplyMove(&state.Game, c.move); err != nil {
				c.reply <- err
				continue
			}
			state.Version++
			state.UpdatedAt = time.Now()
			broadcastChan <- sseEvent{name: "game.updated", data: buildGameSnapshot(&state.Game)}
			for _, robotID := range syncRobotGoals(&state) {
				broadcastChan <- sseEvent{name: "robot.updated", data: buildRobotSnapshot(state, robotID)}
			}
			c.reply <- nil

		case getSnapshotCmd:
			c.reply <- buildAppSnapshot(state)

		case saveCmd:
			c.reply <- saveState(c.filePath, state)
		}
	}
}

func (m *stateManager) newGame() {
	reply := make(chan struct{}, 1)
	m.cmds <- newGameCmd{reply: reply}
	<-reply
}

func (m *stateManager) applyMove(move gameengine.Move) *gameengine.ApplyMoveError {
	reply := make(chan *gameengine.ApplyMoveError, 1)
	m.cmds <- applyMoveCmd{move: move, reply: reply}
	return <-reply
}

func (m *stateManager) getSnapshot() AppSnapshot {
	reply := make(chan AppSnapshot, 1)
	m.cmds <- getSnapshotCmd{reply: reply}
	return <-reply
}

func (m *stateManager) save(filePath string) error {
	reply := make(chan error, 1)
	m.cmds <- saveCmd{filePath: filePath, reply: reply}
	return <-reply
}

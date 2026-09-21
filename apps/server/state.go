package main

import (
	"context"
	"encoding/gob"
	"log"
	"math"
	"os"
	"time"

	"checkerbots/apps/server/fleetapi"
	gameengine "checkerbots/apps/server/game-engine"
	"checkerbots/apps/server/simulator"
)

// appState is the authoritative server state, owned exclusively by the state manager goroutine.
type appState struct {
	Version int64
	Game    gameengine.Game
	// Cache of robot state, separate from FleetController to avoid synchronization issues.
	Robots          map[RobotID]robotInfo
	Assignments     pieceAssignmentState
	Planner         plannerState
	FleetController fleetapi.FleetController
	fleetCmdChan    chan fleetapi.RobotCommand
	robotUpdateChan chan fleetapi.RobotUpdate
	UpdatedAt       time.Time
}

// storedAppState is the disk-serializable form of appState.
type storedAppState struct {
	Version     int64
	Game        gameengine.StoredGame
	Robots      map[RobotID]robotInfo
	Assignments pieceAssignmentState
	Planner     plannerState
	Simulator   simulator.SimState
	UpdatedAt   time.Time
}

// robotInfo holds the live state of a physical robot reported by external systems.
type robotInfo struct {
	Pose      Pose
	UpdatedAt time.Time
}

type pieceAssignmentState struct {
	RobotIDByPieceID map[gameengine.PieceID]RobotID
	PieceIDByRobotID map[RobotID]gameengine.PieceID
}

func (s *pieceAssignmentState) Assign(robotID RobotID, pieceID gameengine.PieceID) {
	s.RobotIDByPieceID[pieceID] = robotID
	s.PieceIDByRobotID[robotID] = pieceID
}

func (s *pieceAssignmentState) Unassign(robotID RobotID) {
	if pieceID, exists := s.PieceIDByRobotID[robotID]; exists {
		delete(s.RobotIDByPieceID, pieceID)
		delete(s.PieceIDByRobotID, robotID)
	}
}

func (s *pieceAssignmentState) GetRobotIDByPieceID(pieceID gameengine.PieceID) (robotID RobotID, ok bool) {
	robotID, ok = s.RobotIDByPieceID[pieceID]
	return
}

func (s *pieceAssignmentState) GetPieceIDByRobotID(robotID RobotID) (pieceID gameengine.PieceID, ok bool) {
	pieceID, ok = s.PieceIDByRobotID[robotID]
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

func makeInitialStoredState() (stored storedAppState) {
	game := gameengine.NewGame8x8()
	stored = storedAppState{
		Game:   game.ToStored(),
		Robots: make(map[RobotID]robotInfo),
		Assignments: pieceAssignmentState{
			RobotIDByPieceID: make(map[gameengine.PieceID]RobotID),
			PieceIDByRobotID: make(map[RobotID]gameengine.PieceID),
		},
		Planner: plannerState{
			Goals: make(map[RobotID]robotGoal),
		},
		Simulator: simulator.NewSimulator(),
		UpdatedAt: time.Now(),
	}

	for _, piece := range stored.Game.Pieces {
		x, y := positionToMM(piece.Position, stored.Game.BoardSize)
		headingRad := float64(0)
		if piece.Side == gameengine.PlayerSideRed {
			headingRad = math.Pi
		}
		if piece.Kind == gameengine.PieceKindKing {
			headingRad = math.Pi - headingRad
		}
		robotID := stored.Simulator.AddRobot(fleetapi.Pose{XMM: x, YMM: y, HeadingRad: headingRad})
		stored.Assignments.Assign(robotID, piece.ID)
	}
	return
}

func (state *appState) ToStored() storedAppState {
	simState, ok := state.FleetController.(*simulator.SimState)
	if !ok {
		log.Panic("fleet controller is not a simulator.SimState")
	}
	return storedAppState{
		Version:     state.Version,
		Game:        state.Game.ToStored(),
		Robots:      state.Robots,
		Assignments: state.Assignments,
		Planner:     state.Planner,
		Simulator:   *simState,
		UpdatedAt:   state.UpdatedAt,
	}
}

func StateFromStored(stored storedAppState) *appState {
	state := appState{
		Version:         stored.Version,
		Game:            gameengine.GameFromStored(stored.Game),
		Robots:          stored.Robots,
		Assignments:     stored.Assignments,
		Planner:         stored.Planner,
		UpdatedAt:       stored.UpdatedAt,
		fleetCmdChan:    make(chan fleetapi.RobotCommand),
		robotUpdateChan: make(chan fleetapi.RobotUpdate),
		FleetController: &stored.Simulator,
	}

	go state.FleetController.Run(
		context.Background(), state.fleetCmdChan, state.robotUpdateChan,
	)
	return &state
}

func saveState(filePath string, stored storedAppState) error {
	if filePath == "" {
		return nil
	}
	log.Printf("saving state to: %s", filePath)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	return gob.NewEncoder(file).Encode(stored)
}

func loadState(filePath string) (stored storedAppState, err error) {
	log.Printf("loading state from: %s", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()
	err = gob.NewDecoder(file).Decode(&stored)
	return
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
func syncRobotGoals(state *appState) {
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
			state.fleetCmdChan <- fleetapi.RobotCommand{
				RobotID:    robotID,
				TargetPose: fleetapi.Pose{XMM: xMM, YMM: yMM},
			}
		}
	}
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
func (m *stateManager) run(initial storedAppState, broadcastChan chan<- sseEvent) {
	state := StateFromStored(initial)
	syncRobotGoals(state)

	for {
		select {
		case cmd, ok := <-m.cmds:
			if !ok {
				return
			}
			switch c := cmd.(type) {
			case newGameCmd:
				state.Game = gameengine.NewGame8x8()
				// TODO: Handle when game piece count changes
				if len(state.Game.Pieces) != len(state.Robots) {
					log.Panicf("piece/robot count mismatch: %d != %d", len(state.Game.Pieces), len(state.Robots))
				}

				state.Version++
				state.UpdatedAt = time.Now()
				syncRobotGoals(state)
				broadcastChan <- sseEvent{name: "game.updated", data: buildGameSnapshot(&state.Game)}
				c.reply <- struct{}{}

			case applyMoveCmd:
				if err := gameengine.ApplyMove(&state.Game, c.move); err != nil {
					c.reply <- err
					continue
				}
				state.Version++
				state.UpdatedAt = time.Now()
				broadcastChan <- sseEvent{name: "game.updated", data: buildGameSnapshot(&state.Game)}
				syncRobotGoals(state)
				c.reply <- nil

			case getSnapshotCmd:
				c.reply <- buildAppSnapshot(state)

			case saveCmd:
				c.reply <- saveState(c.filePath, state.ToStored())
			}
		case update := <-state.robotUpdateChan:
			state.Robots[update.RobotID] = robotInfo{
				Pose: Pose{
					XMM:        update.CurrentPose.XMM,
					YMM:        update.CurrentPose.YMM,
					HeadingRad: update.CurrentPose.HeadingRad,
				},
				UpdatedAt: time.Now(),
			}
			broadcastChan <- sseEvent{name: "robot.updated", data: buildRobotSnapshot(state, update.RobotID)}
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

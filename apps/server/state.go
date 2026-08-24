package main

import (
	"encoding/gob"
	"log"
	"os"
	"time"

	gameengine "checkerbots/apps/server/game-engine"
)

// appState is the authoritative server state, owned exclusively by the state manager goroutine.
type appState struct {
	Version   int64
	Game      gameengine.Game
	Robots    map[string]RobotInfo
	UpdatedAt time.Time
}

// storedAppState is the disk-serializable form of appState.
type storedAppState struct {
	Version   int64
	Game      gameengine.StoredGame
	Robots    map[string]RobotInfo
	UpdatedAt time.Time
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
		Version:   state.Version,
		Game:      gameengine.GameToStored(state.Game),
		Robots:    state.Robots,
		UpdatedAt: state.UpdatedAt,
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
		Version:   stored.Version,
		Game:      gameengine.GameFromStored(stored.Game),
		Robots:    stored.Robots,
		UpdatedAt: stored.UpdatedAt,
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

// syncRobots ensures state.Robots matches the current piece positions.
// It returns all RobotInfo values that were created or updated so the caller
// can publish robot.updated events.
func syncRobots(state *appState) []RobotInfo {
	now := time.Now()
	var changed []RobotInfo

	for _, piece := range state.Game.Pieces {
		xMM, yMM := positionToMM(piece.Position, state.Game.BoardSize)
		pose := Pose{
			XMM:    xMM,
			YMM:    yMM,
			Frame:  CoordinateFrameBoard,
			Source: PoseSourceSimulator,
		}

		existing, exists := state.Robots[piece.ID]
		if !exists {
			// Create a new robot for this piece.
			robot := RobotInfo{
				ID:        piece.ID,
				PieceID:   piece.ID,
				Pose:      pose,
				UpdatedAt: now,
			}
			state.Robots[piece.ID] = robot
			changed = append(changed, robot)
			continue
		}

		if existing.Pose.XMM != pose.XMM || existing.Pose.YMM != pose.YMM {
			existing.Pose = pose
			existing.UpdatedAt = now
			state.Robots[piece.ID] = existing
			changed = append(changed, existing)
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

func newStateManager() *stateManager {
	return &stateManager{cmds: make(chan any)}
}

// run is the state manager's main loop and must be called in its own goroutine.
// publish is called synchronously after each successful state change.
func (m *stateManager) run(initial appState, publish func(sseEvent)) {
	state := initial
	for cmd := range m.cmds {
		switch c := cmd.(type) {
		case newGameCmd:
			state.Game = gameengine.NewGame8x8()
			state.Robots = map[string]RobotInfo{}
			state.Version++
			state.UpdatedAt = time.Now()
			// Send a full snapshot since (right now) all robots move
			// If robot.updated events are sent, the channel buffer will overflow since ~25 events are dumped into it
			syncRobots(&state)
			publish(sseEvent{name: "state.snapshot", data: buildAppSnapshot(state)})
			c.reply <- struct{}{}

		case applyMoveCmd:
			if err := gameengine.ApplyMove(&state.Game, c.move); err != nil {
				c.reply <- err
				continue
			}
			state.Version++
			state.UpdatedAt = time.Now()
			publish(sseEvent{name: "game.updated", data: buildGameSnapshot(&state.Game)})
			for _, robot := range syncRobots(&state) {
				publish(sseEvent{name: "robot.updated", data: robot})
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

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
			state.Version++
			state.UpdatedAt = time.Now()
			publish(sseEvent{name: "game.updated", data: buildGameSnapshot(&state.Game)})
			c.reply <- struct{}{}

		case applyMoveCmd:
			if err := gameengine.ApplyMove(&state.Game, c.move); err != nil {
				c.reply <- err
				continue
			}
			state.Version++
			state.UpdatedAt = time.Now()
			publish(sseEvent{name: "game.updated", data: buildGameSnapshot(&state.Game)})
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

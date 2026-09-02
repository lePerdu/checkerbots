package main

import (
	"encoding/gob"
	"log"
	"math"
	"os"
	"strconv"
	"sync"
	"time"

	gameengine "checkerbots/apps/server/game-engine"
)

// appState is the authoritative server state, owned exclusively by the state manager goroutine.
type appState struct {
	Version         int64
	Game            gameengine.Game
	Robots          map[RobotID]robotInfo
	Assignments     pieceAssignmentState
	Planner         plannerState
	controllers     map[RobotID]robotController
	robotUpdateChan chan robotUpdatedEvent
	controllerWg    sync.WaitGroup
	UpdatedAt       time.Time
}

type robotController struct {
	robotID RobotID
	cmds    chan robotCmd
}

type robotCmd struct {
	CurrentPose Pose
	TargetPose  Pose
}

const ROBOT_TICK_INTERVAL = 200 * time.Millisecond
const ROBOT_SPEED_MMPS = 300.0
const ROBOT_ARRIVED_THRESHOLD_MM = 10.0

type Vec2 struct{ X, Y float64 }

func (v Vec2) Length2() float64 {
	return v.X*v.X + v.Y*v.Y
}

func (v Vec2) Length() float64 {
	return float64(math.Sqrt(float64(v.Length2())))
}

func (v Vec2) Normalize0() Vec2 {
	length := v.Length()
	if length == 0 {
		return v
	}
	return Vec2{X: v.X / length, Y: v.Y / length}
}

func (v Vec2) Scale(s float64) Vec2 {
	return Vec2{X: v.X * s, Y: v.Y * s}
}

type robotUpdatedEvent struct {
	RobotID RobotID
	Pose    Pose
}

func (c *robotController) run(initialPose Pose, updateChan chan<- robotUpdatedEvent) {
	log.Printf("controller: robot %s: starting with initial position: %v", c.robotID, initialPose)
	defer log.Printf("controller: robot %s: stopping", c.robotID)

	curPose := initialPose
	targetPose := initialPose

	ticker := time.NewTicker(ROBOT_TICK_INTERVAL)
	ticker.Stop()
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Printf("controller: robot %s: tick", c.robotID)
			const ROBOT_STEP_MM = ROBOT_SPEED_MMPS * float64(ROBOT_TICK_INTERVAL) / float64(time.Second)
			const ROBOT_STEP_MM_2 = (ROBOT_STEP_MM * ROBOT_STEP_MM)
			const ROBOT_ARRIVED_THRESHOLD_MM_2 = ROBOT_ARRIVED_THRESHOLD_MM * ROBOT_ARRIVED_THRESHOLD_MM

			targetDelta := Vec2{X: targetPose.XMM - curPose.XMM, Y: targetPose.YMM - curPose.YMM}
			dist2 := targetDelta.Length2()
			if dist2 < ROBOT_ARRIVED_THRESHOLD_MM_2 {
				ticker.Stop()
				continue
			}
			stepDist := ROBOT_STEP_MM
			if dist2 < ROBOT_STEP_MM_2 {
				stepDist = math.Sqrt(stepDist)
			}
			stepDelta := targetDelta.Scale(stepDist / math.Sqrt(float64(dist2)))
			curPose.XMM += stepDelta.X
			curPose.YMM += stepDelta.Y
			log.Printf("controller: robot %s: +%v: send update: %v", c.robotID, stepDelta, robotUpdatedEvent{RobotID: c.robotID, Pose: curPose})
			updateChan <- robotUpdatedEvent{RobotID: c.robotID, Pose: curPose}
		case cmd, ok := <-c.cmds:
			if !ok {
				return
			}
			log.Printf("controller: robot %s received command: %v", c.robotID, cmd)
			curPose = cmd.CurrentPose
			targetPose = cmd.TargetPose
			ticker.Reset(ROBOT_TICK_INTERVAL)
		}
	}
}

type RobotID string

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

// storedAppState is the disk-serializable form of appState.
type storedAppState struct {
	Version     int64
	Game        gameengine.StoredGame
	Robots      map[RobotID]robotInfo
	Assignments pieceAssignmentState
	Planner     plannerState
	UpdatedAt   time.Time
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
		UpdatedAt: time.Now(),
	}

	for i, piece := range stored.Game.Pieces {
		robotID := RobotID("r" + strconv.Itoa(i))
		stored.Robots[robotID] = robotInfo{
			// TODO: Figure out initial positions
			Pose:      Pose{},
			UpdatedAt: time.Now(),
		}
		stored.Assignments.Assign(robotID, piece.ID)
	}
	return
}

func (state *appState) ToStored() storedAppState {
	return storedAppState{
		Version:     state.Version,
		Game:        state.Game.ToStored(),
		Robots:      state.Robots,
		Assignments: state.Assignments,
		Planner:     state.Planner,
		UpdatedAt:   state.UpdatedAt,
	}
}

func StateFromStored(stored storedAppState) *appState {
	state := appState{
		Version:     stored.Version,
		Game:        gameengine.GameFromStored(stored.Game),
		Robots:      stored.Robots,
		Assignments: stored.Assignments,
		Planner:     stored.Planner,
		UpdatedAt:   stored.UpdatedAt,
	}
	state.robotUpdateChan = make(chan robotUpdatedEvent, len(state.Robots))
	state.CreateControllers()
	return &state
}

func (state *appState) ClearControllers() {
	for _, c := range state.controllers {
		close(c.cmds)
	}
	state.controllerWg.Wait()
	clear(state.controllers)
}

func (state *appState) CreateControllers() {
	if state.controllers == nil {
		state.controllers = make(map[RobotID]robotController, len(state.Robots))
	}
	state.controllerWg = sync.WaitGroup{}
	for robotID, robot := range state.Robots {
		c := robotController{
			robotID: robotID,
			cmds:    make(chan robotCmd),
		}
		state.controllers[robotID] = c
		state.controllerWg.Go(func() {
			c.run(robot.Pose, state.robotUpdateChan)
		})
	}
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
			state.controllers[robotID].cmds <- robotCmd{
				CurrentPose: state.Robots[robotID].Pose,
				TargetPose:  Pose{XMM: xMM, YMM: yMM},
			}
			log.Printf("send event -> %s: %v", robotID, robotCmd{
				CurrentPose: state.Robots[robotID].Pose,
				TargetPose:  Pose{XMM: xMM, YMM: yMM},
			})
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
				// TODO: Retain same robot objects & controllers across games. Just re-assign and send new commands to move them.
				state.ClearControllers()
				clear(state.Robots)
				clear(state.Assignments.RobotIDByPieceID)
				clear(state.Assignments.PieceIDByRobotID)
				clear(state.Planner.Goals)

				for i, piece := range state.Game.Pieces {
					robotID := RobotID("r" + strconv.Itoa(i))
					state.Robots[robotID] = robotInfo{
						// TODO: Figure out initial positions
						Pose:      Pose{},
						UpdatedAt: time.Now(),
					}
					state.Assignments.Assign(robotID, piece.ID)
				}
				state.CreateControllers()

				state.Version++
				state.UpdatedAt = time.Now()
				// Send a full snapshot since (right now) all robots move
				// If robot.updated events are sent, the channel buffer will overflow since ~25 events are dumped into it
				syncRobotGoals(state)
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
				syncRobotGoals(state)
				c.reply <- nil

			case getSnapshotCmd:
				c.reply <- buildAppSnapshot(state)

			case saveCmd:
				c.reply <- saveState(c.filePath, state.ToStored())
			}
		case update := <-state.robotUpdateChan:
			state.Robots[update.RobotID] = robotInfo{
				Pose:      update.Pose,
				UpdatedAt: time.Now(), // TODO: Include in update event?
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

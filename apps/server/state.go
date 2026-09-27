package main

import (
	"encoding/gob"
	"log"
	"math"
	"os"
	"time"

	"checkerbots/apps/server/fleetapi"
	gameengine "checkerbots/apps/server/game-engine"
)

// appState is the authoritative server state, owned exclusively by the state manager goroutine.
type appState struct {
	Version int64
	Game    gameengine.Game
	// Cache of robot state, separate from FleetController to avoid synchronization issues.
	Robots      map[RobotID]robotInfo
	Assignments pieceAssignmentState
	Planner     plannerState
	UpdatedAt   time.Time

	fleetCmdChan   chan fleetapi.Command
	fleetEventChan chan fleetapi.Event
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

func makeInitialState() appState {
	state := appState{
		Game:   gameengine.NewDefaultGame(),
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
	state.hydrate()
	return state
}

func loadState(filePath string) (state appState, err error) {
	log.Printf("loading state from: %s", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()
	err = gob.NewDecoder(file).Decode(&state)
	state.hydrate()
	return
}

// StateFromStored restores app state from disk. If ctrl is non-nil it is used
// as the FleetController; otherwise the stored simulator state is used.
func (state *appState) hydrate() {
	state.Game.Hydrate()
	state.fleetCmdChan = make(chan fleetapi.Command)
	state.fleetEventChan = make(chan fleetapi.Event)
}

func (state *appState) save(filePath string) error {
	if filePath == "" {
		return nil
	}
	log.Printf("saving state to: %s", filePath)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	return gob.NewEncoder(file).Encode(state)
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

func idlePieceHeading(piece gameengine.Piece) float64 {
	heading := 0.0
	if piece.Side == gameengine.PlayerSideRed {
		heading = math.Pi
	}
	// Kings face the other way
	// TODO: Should this just be the last direction the king moved?
	if piece.Kind == gameengine.PieceKindKing {
		heading = math.Pi - heading
	}
	return heading
}

func targetFleetPose(piece gameengine.Piece, boardSize int) fleetapi.Pose {
	xMM, yMM := positionToMM(piece.Position, boardSize)
	return fleetapi.Pose{
		XMM:        xMM,
		YMM:        yMM,
		HeadingRad: idlePieceHeading(piece),
	}
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
			state.fleetCmdChan <- fleetapi.SetTargetCommand{
				RobotID:    robotID,
				TargetPose: targetFleetPose(piece, state.Game.BoardSize),
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
func (m *stateManager) run(state *appState, broadcastChan chan<- sseEvent) {
	for {
		select {
		case cmd, ok := <-m.cmds:
			if !ok {
				return
			}
			switch c := cmd.(type) {
			case newGameCmd:
				state.Game = gameengine.NewDefaultGame()
				// TODO: Handle when game piece IDs change. Clear and (randomly) re-assign?
				state.Version++
				state.UpdatedAt = time.Now()
				syncRobotGoals(state)
				broadcastChan <- sseEvent{name: "game.updated", data: buildGameSnapshot(&state.Game)}
				c.reply <- struct{}{}

			case applyMoveCmd:
				if err := state.Game.ApplyMove(c.move); err != nil {
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
				c.reply <- state.save(c.filePath)
			}
		case event := <-state.fleetEventChan:
			switch event := event.(type) {
			case fleetapi.RobotConnectedEvent:
				if info, exists := state.Robots[event.RobotID]; exists {
					if event.PoseValid {
						// Use remembered pose
						state.handleRobotPoseUpdate(broadcastChan, event.RobotID, event.CurrentPose)
					} else {
						// Set based on last-known state in server
						state.fleetCmdChan <- fleetapi.SetPoseCommand{
							RobotID: event.RobotID,
							CurrentPose: fleetapi.Pose{
								XMM:        info.Pose.XMM,
								YMM:        info.Pose.YMM,
								HeadingRad: info.Pose.HeadingRad,
							},
						}
						// Clearing the goal is the easiest way to re-sync
						delete(state.Planner.Goals, event.RobotID)
						// TODO: Just sync goal for the new robot
						syncRobotGoals(state)
					}
				} else {
					piece, avail := state.nextUnassignedPiece()
					if !avail {
						log.Printf("no available piece for robot %s", event.RobotID)
						continue
					}
					state.Assignments.Assign(event.RobotID, piece.ID)

					var initialPose fleetapi.Pose
					if event.PoseValid {
						initialPose = event.CurrentPose
					} else {
						// Set based on assigned piece
						initialPose = targetFleetPose(*piece, state.Game.BoardSize)
						state.fleetCmdChan <- fleetapi.SetPoseCommand{
							RobotID:     event.RobotID,
							CurrentPose: initialPose,
						}
					}

					state.Robots[event.RobotID] = robotInfo{
						Pose: Pose{
							XMM:        initialPose.XMM,
							YMM:        initialPose.YMM,
							HeadingRad: initialPose.HeadingRad,
						},
						UpdatedAt: time.Now(),
					}
					// TODO: Just sync goal for the new robot
					syncRobotGoals(state)
					// TODO: New event type
					broadcastChan <- sseEvent{name: "app.snapshot", data: buildAppSnapshot(state)}
				}
			case fleetapi.PoseUpdateEvent:
				state.handleRobotPoseUpdate(broadcastChan, event.RobotID, event.CurrentPose)
			default:
				log.Panic("Unknown event type:", event)
			}
		}
	}
}

func (state *appState) handleRobotPoseUpdate(broadcastChan chan<- sseEvent, robotID RobotID, pose fleetapi.Pose) {
	state.Robots[robotID] = robotInfo{
		Pose: Pose{
			XMM:        pose.XMM,
			YMM:        pose.YMM,
			HeadingRad: pose.HeadingRad,
		},
		UpdatedAt: time.Now(),
	}
	broadcastChan <- sseEvent{name: "robot.updated", data: buildRobotSnapshot(state, robotID)}
}

func (state *appState) nextUnassignedPiece() (piece *gameengine.Piece, ok bool) {
	// TODO: Track state to avoid looping?
	for i := range state.Game.Pieces {
		piece := &state.Game.Pieces[i]
		if _, exists := state.Assignments.GetRobotIDByPieceID(piece.ID); !exists {
			return piece, true
		}
	}
	return nil, false
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

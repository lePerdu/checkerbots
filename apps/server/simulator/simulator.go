package simulator

import (
	"checkerbots/apps/server/fleetapi"
	"checkerbots/apps/server/vec"
	"context"
	"log"
	"strconv"
	"time"
)

type SimState struct {
	Entities          []Entity
	EntityIDByRobotID map[fleetapi.RobotID]EntityID
}

type EntityID int

type Entity struct {
	ID        EntityID
	RobotID   fleetapi.RobotID
	Pos       vec.V2
	Vel       vec.V2
	TargetPos vec.V2
}

const ENTITY_RADIUS = 0.34 / 2.0
const ROBOT_MAX_SPEED = 0.3
const ROBOT_TOLERANCE = 0.02
const SIM_TIME_STEP = time.Second / 30.0

func NewSimulator() SimState {
	return SimState{
		EntityIDByRobotID: make(map[fleetapi.RobotID]EntityID),
	}
}

func (s *SimState) AddRobot(initialPose fleetapi.Pose) fleetapi.RobotID {
	entityID := EntityID(len(s.Entities))
	robotID := fleetapi.RobotID("r" + strconv.Itoa(int(entityID)))
	pos := vec.V2{X: initialPose.XMM, Y: initialPose.YMM}.Scale(1.0 / 1000.0)
	s.Entities = append(s.Entities, Entity{ID: entityID, RobotID: robotID, Pos: pos})
	s.EntityIDByRobotID[robotID] = entityID
	return robotID
}

func (s *SimState) Run(
	ctx context.Context,
	cmdChan <-chan fleetapi.RobotCommand,
	updateChan chan<- fleetapi.RobotUpdate,
) {
	ticker := time.NewTicker(SIM_TIME_STEP)
	defer ticker.Stop()

	lastUpdate := time.Now()
	for {
		select {
		case t := <-ticker.C:
			for ; lastUpdate.Before(t); lastUpdate = lastUpdate.Add(SIM_TIME_STEP) {
				s.step(SIM_TIME_STEP, updateChan)
			}
		case cmd, ok := <-cmdChan:
			if !ok {
				return
			}
			s.handleCommand(cmd, updateChan)
		case <-ctx.Done():
			return
		}
	}
}

func (s *SimState) handleCommand(
	cmd fleetapi.RobotCommand, updateChan chan<- fleetapi.RobotUpdate,
) {
	_ = updateChan
	entity_id, exists := s.EntityIDByRobotID[cmd.RobotID]
	if !exists {
		log.Panic("sim: unknown robot ID:", cmd.RobotID)
	}
	s.Entities[entity_id].TargetPos = vec.V2{
		X: cmd.TargetPose.XMM, Y: cmd.TargetPose.YMM,
	}.Scale(1.0 / 1000.0)
}

func (s *SimState) step(dt time.Duration, updateChan chan<- fleetapi.RobotUpdate) {
	for entityID := range s.Entities {
		s.stepEntity(&s.Entities[entityID], dt, updateChan)
	}
}

func (s *SimState) stepEntity(
	entity *Entity, dt time.Duration, updateChan chan<- fleetapi.RobotUpdate,
) {
	targetDelta := entity.TargetPos.Sub(entity.Pos)
	targetDist := targetDelta.Length()
	if targetDist < ROBOT_TOLERANCE {
		entity.Vel = vec.V2{}
		return
	}
	entity.Vel = targetDelta.Normalize().Scale(ROBOT_MAX_SPEED)
	stepDelta := entity.Vel.Scale(dt.Seconds())
	if stepDelta.Length() > targetDist {
		stepDelta = targetDelta
	}
	entity.Pos = entity.Pos.Add(stepDelta)
	updateChan <- fleetapi.RobotUpdate{
		RobotID: entity.RobotID,
		CurrentPose: fleetapi.Pose{
			XMM: entity.Pos.X * 1000.0,
			YMM: entity.Pos.Y * 1000.0,
		},
	}
}

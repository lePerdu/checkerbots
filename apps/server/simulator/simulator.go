package simulator

import (
	"checkerbots/apps/server/fleetapi"
	"checkerbots/apps/server/vec"
	"context"
	"log"
	"strconv"
	"time"
)

// Fixed-sized ring buffer
// TODO: Move this to shared package
type ring[T any] struct {
	items    []T
	off, len int
}

func NewRing[T any](cap int) ring[T] {
	return ring[T]{
		items: make([]T, cap),
	}
}

func (r *ring[T]) Cap() int {
	return len(r.items)
}

func (r *ring[T]) Len() int {
	return r.len
}

func (r *ring[T]) PushBack(item T) {
	cap := len(r.items)
	if r.len == cap {
		panic("ring buffer full")
	}
	r.items[(r.off+r.len)%cap] = item
	r.len += 1
}

func (r *ring[T]) PushFront(item T) {
	cap := len(r.items)
	if r.len == cap {
		panic("ring buffer full")
	}
	r.off = (r.off - 1 + cap) % cap
	r.items[r.off] = item
	r.len += 1
}

func (r *ring[T]) PopBack() (item T, ok bool) {
	cap := len(r.items)
	if r.len == 0 {
		return
	}
	r.len -= 1
	item = r.items[(r.off+r.len)%cap]
	ok = true
	return
}

func (r *ring[T]) PopFront() (item T, ok bool) {
	cap := len(r.items)
	if r.len == 0 {
		return
	}
	r.len -= 1
	item = r.items[r.off]
	ok = true
	r.off = (r.off + 1) % cap
	return
}

type SimState struct {
	Entities            []Entity
	EntityIDByRobotID   map[fleetapi.RobotID]EntityID
	updateRequiredQueue ring[EntityID]
}

type EntityID int

type Entity struct {
	ID        EntityID
	RobotID   fleetapi.RobotID
	Pos       vec.V2
	Vel       vec.V2
	TargetPos vec.V2
	// Whether the latest entity state has been published to the update channel
	needsUpdateQueued bool
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

type EntityUpdate struct {
	fleetapi.RobotUpdate
	EntityID EntityID
}

func (s *SimState) Run(
	ctx context.Context,
	cmdChan <-chan fleetapi.RobotCommand,
	updateChan chan<- fleetapi.RobotUpdate,
) {
	s.updateRequiredQueue = NewRing[EntityID](len(s.Entities))
	for id := range s.Entities {
		s.Entities[id].needsUpdateQueued = false
		s.updateRequiredQueue.PushBack(EntityID(id))
	}

	ticker := time.NewTicker(SIM_TIME_STEP)
	defer ticker.Stop()

	lastUpdate := time.Now()

	entityUpdatePending := false
	var nextEntityUpdate fleetapi.RobotUpdate

	for {
		if !entityUpdatePending && s.updateRequiredQueue.Len() > 0 {
			entityID, _ := s.updateRequiredQueue.PopFront()
			nextEntityUpdate = s.makeRobotUpdate(entityID)
			s.Entities[entityID].needsUpdateQueued = true
			entityUpdatePending = true
		}

		// Make this nil so that it will be skipped in the select if no update is pending
		var optUpdateChan chan<- fleetapi.RobotUpdate
		if entityUpdatePending {
			optUpdateChan = updateChan
		}

		select {
		case t := <-ticker.C:
			for ; lastUpdate.Before(t); lastUpdate = lastUpdate.Add(SIM_TIME_STEP) {
				s.step(SIM_TIME_STEP)
			}
		case cmd, ok := <-cmdChan:
			if !ok {
				return
			}
			s.handleCommand(cmd)
		case optUpdateChan <- nextEntityUpdate:
			entityUpdatePending = false
		case <-ctx.Done():
			return
		}
	}
}

func (s *SimState) handleCommand(cmd fleetapi.RobotCommand) {
	entity_id, exists := s.EntityIDByRobotID[cmd.RobotID]
	if !exists {
		log.Panic("sim: unknown robot ID:", cmd.RobotID)
	}
	s.Entities[entity_id].TargetPos = vec.V2{
		X: cmd.TargetPose.XMM, Y: cmd.TargetPose.YMM,
	}.Scale(1.0 / 1000.0)
}

func (s *SimState) step(dt time.Duration) {
	for entityID := range s.Entities {
		s.stepEntity(&s.Entities[entityID], dt)
	}
	s.resolveEntityCollisions()
}

func (s *SimState) resolveEntityCollisions() {
	minDist := ENTITY_RADIUS * 2.0
	minDist2 := minDist * minDist

	for i := range s.Entities {
		for j := i + 1; j < len(s.Entities); j++ {
			a := &s.Entities[i]
			b := &s.Entities[j]

			delta := b.Pos.Sub(a.Pos)
			dist2 := delta.Length2()
			if dist2 >= minDist2 {
				continue
			}

			normal := vec.V2{X: 1}
			dist := 0.0
			if dist2 > 0 {
				dist = delta.Length()
				normal = delta.Scale(1.0 / dist)
			}

			correction := normal.Scale((minDist - dist) / 2.0)
			a.Pos = a.Pos.Sub(correction)
			b.Pos = b.Pos.Add(correction)

			if a.needsUpdateQueued {
				a.needsUpdateQueued = false
				s.updateRequiredQueue.PushBack(a.ID)
			}
			if b.needsUpdateQueued {
				b.needsUpdateQueued = false
				s.updateRequiredQueue.PushBack(b.ID)
			}
		}
	}
}

func (s *SimState) makeRobotUpdate(id EntityID) fleetapi.RobotUpdate {
	entity := s.Entities[id]
	return fleetapi.RobotUpdate{
		RobotID: entity.RobotID,
		CurrentPose: fleetapi.Pose{
			XMM: entity.Pos.X * 1000.0,
			YMM: entity.Pos.Y * 1000.0,
		},
	}
}

func (s *SimState) stepEntity(entity *Entity, dt time.Duration) {
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
	if entity.needsUpdateQueued {
		entity.needsUpdateQueued = false
		s.updateRequiredQueue.PushBack(entity.ID)
	}
}

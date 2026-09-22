package simulator

import (
	"checkerbots/apps/server/fleetapi"
	"checkerbots/apps/server/vec"
	"context"
	"log"
	"math"
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

type Pose struct {
	Pos vec.V2
	// Heading is the facing direction in radians.
	// 0 points from the black side toward the red side (+Y). Preserved from the
	// last non-zero velocity so the robot keeps its last heading when stopped.
	Heading float64
}

func poseFromFleetPose(pose fleetapi.Pose) Pose {
	return Pose{
		Pos:     vec.V2{X: pose.XMM, Y: pose.YMM}.Scale(1.0 / 1000.0),
		Heading: pose.HeadingRad,
	}
}

func poseToFleetPose(pose Pose) fleetapi.Pose {
	return fleetapi.Pose{
		XMM:        pose.Pos.X * 1000.0,
		YMM:        pose.Pos.Y * 1000.0,
		HeadingRad: pose.Heading,
	}
}

type Entity struct {
	ID            EntityID
	RobotID       fleetapi.RobotID
	Pose          Pose
	TargetPose    Pose
	reachedTarget bool
	// Whether the latest entity state has been published to the update channel
	needsUpdateQueued bool
}

const ENTITY_RADIUS = 0.34 / 2.0
const ROBOT_WHEEL_BASE_RADIUS = 0.232 / 2.0
const ROBOT_MAX_SPEED = 0.3
const ROBOT_POS_TOLERANCE = 0.02
const ROBOT_MAX_ANGULAR_SPEED = math.Pi / 2.0
const ROBOT_ANGLE_TOLERANCE = math.Pi / 180.0
const SIM_TIME_STEP = time.Second / 30.0

func NewSimulator() SimState {
	return SimState{
		EntityIDByRobotID: make(map[fleetapi.RobotID]EntityID),
	}
}

func (s *SimState) AddRobot(initialPose fleetapi.Pose) fleetapi.RobotID {
	entityID := EntityID(len(s.Entities))
	robotID := fleetapi.RobotID("r" + strconv.Itoa(int(entityID)))
	s.Entities = append(s.Entities, Entity{
		ID:      entityID,
		RobotID: robotID,
		Pose:    poseFromFleetPose(initialPose),
	})
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
	s.Entities[entity_id].TargetPose = poseFromFleetPose(cmd.TargetPose)
}

func (s *SimState) step(dt time.Duration) {
	for entityID := range s.Entities {
		entity := &s.Entities[entityID]
		prevPose := entity.Pose
		entity.step(dt)
		if prevPose != entity.Pose && entity.needsUpdateQueued {
			entity.needsUpdateQueued = false
			s.updateRequiredQueue.PushBack(entity.ID)
		}
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

			delta := b.Pose.Pos.Sub(a.Pose.Pos)
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
			a.Pose.Pos = a.Pose.Pos.Sub(correction)
			b.Pose.Pos = b.Pose.Pos.Add(correction)

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
		RobotID:     entity.RobotID,
		CurrentPose: poseToFleetPose(entity.Pose),
	}
}

func (entity *Entity) step(dt time.Duration) {
	entity.stepCurved(dt)
}

func (entity *Entity) stepCurved(dt time.Duration) {
	targetDelta := entity.TargetPose.Pos.Sub(entity.Pose.Pos)
	targetDist := targetDelta.Length()
	if targetDist < ROBOT_POS_TOLERANCE {
		entity.stepHeading(entity.TargetPose.Heading, dt)
		return
	}

	headingToTargetPos := math.Atan2(targetDelta.X, targetDelta.Y)
	deltaHeading := getHeadingDelta(headingToTargetPos - entity.Pose.Heading)
	// Turn until facing the general direction
	if math.Abs(deltaHeading) > math.Pi/6 {
		entity.stepHeading(headingToTargetPos, dt)
		return
	}

	// Go straight if radius would be <= 1mm
	linearTol := math.Asin(0.001 * 2 / targetDist)
	if math.Abs(deltaHeading) <= linearTol {
		vel := targetDelta.Normalize().Scale(ROBOT_MAX_SPEED)
		// Update heading from velocity. 0 = +Y (toward red side); increases clockwise.
		stepDelta := vel.Scale(dt.Seconds())
		if stepDelta.Length() > targetDist {
			stepDelta = targetDelta
		}
		entity.Pose.Pos = entity.Pose.Pos.Add(stepDelta)
		return
	}

	// Negative = turn left, positive = turn right
	radius := targetDist / 2 / math.Sin(deltaHeading)

	k := ROBOT_WHEEL_BASE_RADIUS / 2 / radius
	ratioVrVl := (1 + k) / (1 - k)
	var vl, vr float64
	// Whichever side is greater must be positive so that the linear speed is positive
	if math.Abs(ratioVrVl) >= 1 {
		vr = ROBOT_MAX_SPEED
		vl = vr / ratioVrVl
	} else {
		vl = ROBOT_MAX_SPEED
		vr = vl * ratioVrVl
	}

	linearSpeed := (vr + vl) / 2
	if linearSpeed < 0 {
		log.Panicf("Unexpected negative linear speed: %f: radius=%f vr=%f vl=%f", linearSpeed, radius, vr, vl)
	}
	if linearSpeed*dt.Seconds() > targetDist {
		linearSpeed = targetDist / dt.Seconds()
	}

	linearVel := vec.V2{
		// sin/cos are flipped since heading is +Y->-X, not +X->+Y
		X: math.Sin(entity.Pose.Heading),
		Y: math.Cos(entity.Pose.Heading),
	}.Scale(linearSpeed)
	angularVel := linearSpeed / radius

	// TODO: Slow down as we approach the target
	entity.Pose.Pos = entity.Pose.Pos.Add(linearVel.Scale(dt.Seconds()))
	entity.Pose.Heading += angularVel * dt.Seconds()
}

func (entity *Entity) stepLinear(dt time.Duration) {
	targetDelta := entity.TargetPose.Pos.Sub(entity.Pose.Pos)
	targetDist := targetDelta.Length()
	if targetDist < ROBOT_POS_TOLERANCE {
		entity.stepHeading(entity.TargetPose.Heading, dt)
		return
	}

	// Fix heading first
	headingToTargetPos := math.Atan2(targetDelta.X, targetDelta.Y)
	if !entity.stepHeading(headingToTargetPos, dt) {
		return
	}

	vel := targetDelta.Normalize().Scale(ROBOT_MAX_SPEED)
	// Update heading from velocity. 0 = +Y (toward red side); increases clockwise.
	stepDelta := vel.Scale(dt.Seconds())
	if stepDelta.Length() > targetDist {
		stepDelta = targetDelta
	}
	entity.Pose.Pos = entity.Pose.Pos.Add(stepDelta)
}

func (entity *Entity) stepHeading(targetHeading float64, dt time.Duration) (done bool) {
	headingDelta := getHeadingDelta(targetHeading - entity.Pose.Heading)
	if math.Abs(headingDelta) < ROBOT_ANGLE_TOLERANCE {
		return true
	}

	headingStepMag := min(ROBOT_MAX_ANGULAR_SPEED*dt.Seconds(), math.Abs(headingDelta))
	headingStepDelta := math.Copysign(headingStepMag, headingDelta)
	entity.Pose.Heading += headingStepDelta
	return false
}

// getHeadingDelta returns the shortest angular distance between two headings, in the range [-pi, pi]
func getHeadingDelta(delta float64) float64 {
	delta = math.Mod(delta, 2*math.Pi)
	if delta > math.Pi {
		delta -= 2 * math.Pi
	}
	if delta < -math.Pi {
		delta += 2 * math.Pi
	}
	return delta
}

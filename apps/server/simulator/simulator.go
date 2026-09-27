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

func MakeRing[T any](cap int) ring[T] {
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
	entities            []entity
	entityIDByRobotID   map[fleetapi.RobotID]entityID
	updateRequiredQueue ring[entityID]
}

type entityID int

type entityPose struct {
	Pos vec.V2
	// Heading is the facing direction in radians.
	// 0 points from the black side toward the red side (+Y). Preserved from the
	// last non-zero velocity so the robot keeps its last heading when stopped.
	Heading float64
}

func poseFromFleetPose(pose fleetapi.Pose) entityPose {
	return entityPose{
		Pos:     vec.V2{X: pose.XMM, Y: pose.YMM}.Scale(1.0 / 1000.0),
		Heading: pose.HeadingRad,
	}
}

func poseToFleetPose(pose entityPose) fleetapi.Pose {
	return fleetapi.Pose{
		XMM:        pose.Pos.X * 1000.0,
		YMM:        pose.Pos.Y * 1000.0,
		HeadingRad: pose.Heading,
	}
}

type entity struct {
	id              entityID
	robotID         fleetapi.RobotID
	pose            entityPose
	targetPose      entityPose
	reachedTarget   bool
	hasUpdateQueued bool
}

const ENTITY_RADIUS = 0.34 / 2.0
const ROBOT_WHEEL_BASE_RADIUS = 0.232 / 2.0
const ROBOT_MAX_SPEED = 0.3
const ROBOT_POS_TOLERANCE = 0.02
const ROBOT_MAX_ANGULAR_SPEED = math.Pi / 2.0
const ROBOT_ANGLE_TOLERANCE = math.Pi / 180.0
const SIM_TIME_STEP = time.Second / 30.0

func NewSimulator(numRobots int) *SimState {
	state := &SimState{
		entityIDByRobotID:   make(map[fleetapi.RobotID]entityID),
		entities:            make([]entity, 0, numRobots),
		updateRequiredQueue: MakeRing[entityID](numRobots),
	}
	for range numRobots {
		state.addRobot(fleetapi.Pose{})
	}
	return state
}

func (s *SimState) addRobot(initialPose fleetapi.Pose) fleetapi.RobotID {
	entityID := entityID(len(s.entities))
	robotID := fleetapi.RobotID("r" + strconv.Itoa(int(entityID)))
	s.entities = append(s.entities, entity{
		id:            entityID,
		robotID:       robotID,
		pose:          poseFromFleetPose(initialPose),
		targetPose:    poseFromFleetPose(initialPose),
		reachedTarget: true,
	})
	s.entityIDByRobotID[robotID] = entityID
	return robotID
}

type entityUpdate struct {
	fleetapi.PoseUpdateEvent
	entityID entityID
}

func (s *SimState) Run(
	ctx context.Context,
	cmdChan <-chan fleetapi.Command,
	eventChan chan<- fleetapi.Event,
) {
	// Simple way to send connected events concurrently with receiving commands
	// These events don't need to be synchronized, since:
	// - Commands for individual robots won't arrive until after connected message is processed
	// - Updates for individual robots won't be sent until commands are processed
	go func() {
		for _, entity := range s.entities {
			eventChan <- fleetapi.RobotConnectedEvent{
				RobotID:     entity.robotID,
				CurrentPose: fleetapi.Pose{},
				PoseValid:   false,
			}
		}
	}()

	ticker := time.NewTicker(SIM_TIME_STEP)
	defer ticker.Stop()

	lastUpdate := time.Now()

	entityUpdatePending := false
	var nextEntityUpdate fleetapi.PoseUpdateEvent

	for {
		if !entityUpdatePending && s.updateRequiredQueue.Len() > 0 {
			entityID, _ := s.updateRequiredQueue.PopFront()
			nextEntityUpdate = s.makeRobotUpdate(entityID)
			s.entities[entityID].hasUpdateQueued = false
			entityUpdatePending = true
		}

		// Make this nil so that it will be skipped in the select if no update is pending
		var optEventChan chan<- fleetapi.Event
		if entityUpdatePending {
			optEventChan = eventChan
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
		case optEventChan <- nextEntityUpdate:
			entityUpdatePending = false
		case <-ctx.Done():
			return
		}
	}
}

func (s *SimState) handleCommand(cmd fleetapi.Command) {
	switch cmd := cmd.(type) {
	case fleetapi.SetPoseCommand:
		entityID, exists := s.entityIDByRobotID[cmd.RobotID]
		if !exists {
			log.Print("sim: unknown robot ID:", cmd.RobotID)
			return
		}
		entity := &s.entities[entityID]
		// Manual pose setting shouldn't happen often, so don't bother diffing the pose
		entity.pose = poseFromFleetPose(cmd.CurrentPose)
		s.handleEntityUpdated(entity)
	case fleetapi.SetTargetCommand:
		entityID, exists := s.entityIDByRobotID[cmd.RobotID]
		if !exists {
			log.Print("sim: unknown robot ID:", cmd.RobotID)
			return
		}
		s.entities[entityID].targetPose = poseFromFleetPose(cmd.TargetPose)
	default:
		log.Panic("Unknown command type:", cmd)
	}
}

func (s *SimState) handleEntityUpdated(entity *entity) {
	if !entity.hasUpdateQueued {
		s.updateRequiredQueue.PushBack(entity.id)
		entity.hasUpdateQueued = true
	}
}

func (s *SimState) step(dt time.Duration) {
	for entityID := range s.entities {
		entity := &s.entities[entityID]
		prevPose := entity.pose
		entity.step(dt)
		if prevPose != entity.pose {
			s.handleEntityUpdated(entity)
		}
	}
	s.resolveEntityCollisions()
}

func (s *SimState) resolveEntityCollisions() {
	minDist := ENTITY_RADIUS * 2.0
	minDist2 := minDist * minDist

	for i := range s.entities {
		for j := i + 1; j < len(s.entities); j++ {
			a := &s.entities[i]
			b := &s.entities[j]

			delta := b.pose.Pos.Sub(a.pose.Pos)
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
			a.pose.Pos = a.pose.Pos.Sub(correction)
			b.pose.Pos = b.pose.Pos.Add(correction)

			s.handleEntityUpdated(a)
			s.handleEntityUpdated(b)
		}
	}
}

func (s *SimState) makeRobotUpdate(id entityID) fleetapi.PoseUpdateEvent {
	entity := s.entities[id]
	return fleetapi.PoseUpdateEvent{
		RobotID:     entity.robotID,
		CurrentPose: poseToFleetPose(entity.pose),
	}
}

func (entity *entity) step(dt time.Duration) {
	entity.stepCurved(dt)
}

func (entity *entity) stepCurved(dt time.Duration) {
	targetDelta := entity.targetPose.Pos.Sub(entity.pose.Pos)
	targetDist := targetDelta.Length()
	if targetDist < ROBOT_POS_TOLERANCE {
		entity.stepHeading(entity.targetPose.Heading, dt)
		return
	}

	headingToTargetPos := math.Atan2(targetDelta.X, targetDelta.Y)
	deltaHeading := getHeadingDelta(headingToTargetPos - entity.pose.Heading)
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
		entity.pose.Pos = entity.pose.Pos.Add(stepDelta)
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
		X: math.Sin(entity.pose.Heading),
		Y: math.Cos(entity.pose.Heading),
	}.Scale(linearSpeed)
	angularVel := linearSpeed / radius

	// TODO: Slow down as we approach the target
	entity.pose.Pos = entity.pose.Pos.Add(linearVel.Scale(dt.Seconds()))
	entity.pose.Heading += angularVel * dt.Seconds()
}

func (entity *entity) stepLinear(dt time.Duration) {
	targetDelta := entity.targetPose.Pos.Sub(entity.pose.Pos)
	targetDist := targetDelta.Length()
	if targetDist < ROBOT_POS_TOLERANCE {
		entity.stepHeading(entity.targetPose.Heading, dt)
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
	entity.pose.Pos = entity.pose.Pos.Add(stepDelta)
}

func (entity *entity) stepHeading(targetHeading float64, dt time.Duration) (done bool) {
	headingDelta := getHeadingDelta(targetHeading - entity.pose.Heading)
	if math.Abs(headingDelta) < ROBOT_ANGLE_TOLERANCE {
		return true
	}

	headingStepMag := min(ROBOT_MAX_ANGULAR_SPEED*dt.Seconds(), math.Abs(headingDelta))
	headingStepDelta := math.Copysign(headingStepMag, headingDelta)
	entity.pose.Heading += headingStepDelta
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

// sim-robot connects to the coordination server as a simulated robot, implementing
// the robo-proto WebSocket API. It accepts SetTargetCommand / SetPoseCommand from
// the server and simulates curved robot motion, sending PoseUpdateMessage as the
// pose changes. Collision detection is not implemented.
//
// On disconnect, sim-robot reconnects automatically using exponential backoff
// starting at 1 s and capped at 10 s. Physics state (pose, target) is preserved
// across reconnections.
//
// Usage:
//
//	WS_ENDPOINT=ws://localhost:8080/api/robots/ws sim-robot ROBOT_ID
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/coder/websocket"

	roboproto "checkerbots/packages/robo-proto"
)

// Physics constants match the server simulator.
const (
	robotWheelBaseRadius = 0.232 / 2.0
	robotMaxSpeed        = 0.3
	robotPosTolerance    = 0.02
	robotMaxAngularSpeed = math.Pi / 2.0
	robotAngleTolerance  = math.Pi / 180.0
	simTimeStep          = time.Second / 30
	heartbeatInterval    = 1 * time.Second
)

const (
	backoffInitial = 1 * time.Second
	backoffMax     = 10 * time.Second
)

var msgSeq atomic.Int64

func newMsgID() string {
	return fmt.Sprintf("msg-%d", msgSeq.Add(1))
}

// v2 is an inline 2D vector type so the sim-robot has no cross-app dependency.
type v2 struct{ x, y float64 }

func (v v2) add(u v2) v2        { return v2{v.x + u.x, v.y + u.y} }
func (v v2) sub(u v2) v2        { return v2{v.x - u.x, v.y - u.y} }
func (v v2) scale(s float64) v2 { return v2{v.x * s, v.y * s} }
func (v v2) length2() float64   { return v.x*v.x + v.y*v.y }
func (v v2) length() float64    { return math.Sqrt(v.length2()) }
func (v v2) normalize() v2      { return v.scale(1.0 / v.length()) }

type simPose struct {
	pos     v2
	heading float64 // radians; 0 = +Y (toward red side), increases clockwise
}

func poseFromProto(p roboproto.Pose) simPose {
	return simPose{
		pos:     v2{x: p.XMM / 1000.0, y: p.YMM / 1000.0},
		heading: p.HeadingRad,
	}
}

func poseToProto(p simPose) roboproto.Pose {
	return roboproto.Pose{
		Pos:        roboproto.Pos{XMM: p.pos.x * 1000.0, YMM: p.pos.y * 1000.0},
		HeadingRad: p.heading,
	}
}

// robotState holds physics state that persists across reconnections.
type robotState struct {
	current   simPose
	target    simPose
	poseDirty bool // true when current pose should be sent on the next tick
	poseValid bool // true when the pose has been set by the server (SetPoseCommand)
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s ROBOT_ID", os.Args[0])
	}
	robotID := os.Args[1]

	endpoint := os.Getenv("WS_ENDPOINT")
	if endpoint == "" {
		log.Fatal("WS_ENDPOINT environment variable is required")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	run(ctx, robotID, endpoint)
}

// run connects and reconnects forever until ctx is cancelled, applying
// exponential backoff (1 s → 10 s) between attempts. Physics state is
// preserved so the robot remembers its pose after a reconnect.
func run(ctx context.Context, robotID, endpoint string) {
	state := robotState{poseDirty: true}
	backoff := backoffInitial

	for {
		err := connect(ctx, robotID, endpoint, &state, &backoff)
		if ctx.Err() != nil {
			return
		}
		log.Printf("sim-robot %s: disconnected (%v), reconnecting in %v", robotID, err, backoff)
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return
		}
		backoff = min(backoff*2, backoffMax)
	}
}

// connect dials the server, sends HelloMessage, then runs the message loop
// until the connection drops or ctx is cancelled. It resets *backoff to the
// initial value once the dial succeeds, so a transient mid-session drop
// retries quickly rather than continuing a long backoff sequence.
func connect(ctx context.Context, robotID, endpoint string, state *robotState, backoff *time.Duration) error {
	conn, _, err := websocket.Dial(ctx, endpoint, nil)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.CloseNow()
	*backoff = backoffInitial // reset: a successful dial means we're talking to the server

	hello := roboproto.HelloMessage{
		Envelope: roboproto.Envelope{
			ProtocolVersion: roboproto.ProtocolV0,
			Type:            roboproto.MessageTypeHello,
			MessageID:       newMsgID(),
			RobotID:         robotID,
			Timestamp:       time.Now().UTC(),
		},
		RobotModel:      "sim-robot",
		DisplayName:     robotID,
		SoftwareVersion: "dev",
		Status:          roboproto.RobotStatusIdle,
		Pose:            poseToProto(state.current),
		PoseValid:       state.poseValid,
	}
	data, err := json.Marshal(hello)
	if err != nil {
		return fmt.Errorf("marshal hello: %w", err)
	}
	if err := conn.Write(ctx, websocket.MessageText, data); err != nil {
		return fmt.Errorf("write hello: %w", err)
	}
	log.Printf("sim-robot %s: connected to %s", robotID, endpoint)

	// Resend current pose immediately so the server is in sync.
	state.poseDirty = true

	// Read incoming messages in a goroutine and forward them to the select loop.
	type incomingMsg struct {
		data []byte
		err  error
	}
	msgCh := make(chan incomingMsg, 8)
	go func() {
		for {
			_, raw, err := conn.Read(ctx)
			select {
			case msgCh <- incomingMsg{data: raw, err: err}:
			case <-ctx.Done():
				return
			}
			if err != nil {
				return
			}
		}
	}()

	simTicker := time.NewTicker(simTimeStep)
	defer simTicker.Stop()
	hbTicker := time.NewTicker(heartbeatInterval)
	defer hbTicker.Stop()

	// Initialise lastStep to now to avoid a step catch-up burst after a long
	// disconnect; the simulated robot simply freezes while offline.
	lastStep := time.Now()

	for {
		select {
		case <-ctx.Done():
			conn.Close(websocket.StatusNormalClosure, "shutdown")
			return nil

		case msg := <-msgCh:
			if msg.err != nil {
				return fmt.Errorf("read: %w", msg.err)
			}
			var env roboproto.Envelope
			if err := json.Unmarshal(msg.data, &env); err != nil {
				log.Printf("sim-robot %s: bad envelope: %v", robotID, err)
				continue
			}
			switch env.Type {
			case roboproto.MessageTypeSetTarget:
				var cmd roboproto.SetTargetCommand
				if err := json.Unmarshal(msg.data, &cmd); err != nil {
					log.Printf("sim-robot %s: parse set_target: %v", robotID, err)
					continue
				}
				state.target = poseFromProto(cmd.Pose)
				log.Printf("sim-robot %s: set_target pos=(%.1f,%.1f)mm heading=%.3frad",
					robotID, cmd.Pose.XMM, cmd.Pose.YMM, cmd.Pose.HeadingRad)
			case roboproto.MessageTypeSetPose:
				var cmd roboproto.SetPoseCommand
				if err := json.Unmarshal(msg.data, &cmd); err != nil {
					log.Printf("sim-robot %s: parse set_pose: %v", robotID, err)
					continue
				}
				state.current = poseFromProto(cmd.Pose)
				state.target = state.current
				state.poseDirty = true
				state.poseValid = true
				log.Printf("sim-robot %s: set_pose pos=(%.1f,%.1f)mm heading=%.3frad",
					robotID, cmd.Pose.XMM, cmd.Pose.YMM, cmd.Pose.HeadingRad)
			case roboproto.MessageTypeStop:
				state.target = state.current
				log.Printf("sim-robot %s: stop", robotID)
			case roboproto.MessageTypePing:
				pong := roboproto.Envelope{
					ProtocolVersion: roboproto.ProtocolV0,
					Type:            roboproto.MessageTypePong,
					MessageID:       newMsgID(),
					RobotID:         robotID,
					Timestamp:       time.Now().UTC(),
				}
				if data, err := json.Marshal(pong); err == nil {
					conn.Write(ctx, websocket.MessageText, data) //nolint:errcheck
				}
			default:
				log.Printf("sim-robot %s: unhandled message type %q", robotID, env.Type)
			}

		case t := <-simTicker.C:
			// Catch up in fixed-size steps to keep physics deterministic.
			for ; lastStep.Before(t); lastStep = lastStep.Add(simTimeStep) {
				prev := state.current
				stepCurved(&state.current, state.target, simTimeStep)
				if state.current != prev {
					state.poseDirty = true
				}
			}
			if state.poseDirty {
				if err := sendPoseUpdate(ctx, conn, robotID, state.current); err != nil {
					return fmt.Errorf("send pose update: %w", err)
				}
				state.poseDirty = false
			}

		case <-hbTicker.C:
			if err := sendHeartbeat(ctx, conn, robotID); err != nil {
				return fmt.Errorf("send heartbeat: %w", err)
			}
		}
	}
}

func sendPoseUpdate(ctx context.Context, conn *websocket.Conn, robotID string, p simPose) error {
	msg := roboproto.PoseUpdateMessage{
		Envelope: roboproto.Envelope{
			ProtocolVersion: roboproto.ProtocolV0,
			Type:            roboproto.MessageTypePoseUpdate,
			MessageID:       newMsgID(),
			RobotID:         robotID,
			Timestamp:       time.Now().UTC(),
		},
		Pose:       poseToProto(p),
		Source:     roboproto.PoseSourceSimulator,
		Confidence: 1.0,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return conn.Write(ctx, websocket.MessageText, data)
}

func sendHeartbeat(ctx context.Context, conn *websocket.Conn, robotID string) error {
	msg := roboproto.HeartbeatMessage{
		Envelope: roboproto.Envelope{
			ProtocolVersion: roboproto.ProtocolV0,
			Type:            roboproto.MessageTypeHeartbeat,
			MessageID:       newMsgID(),
			RobotID:         robotID,
			Timestamp:       time.Now().UTC(),
		},
		Status: roboproto.RobotStatusIdle,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return conn.Write(ctx, websocket.MessageText, data)
}

// stepCurved advances current pose toward target by dt using curved-motion kinematics.
// This mirrors the server simulator's stepCurved algorithm exactly.
func stepCurved(current *simPose, target simPose, dt time.Duration) {
	targetDelta := target.pos.sub(current.pos)
	targetDist := targetDelta.length()
	if targetDist < robotPosTolerance {
		stepHeading(current, target.heading, dt)
		return
	}

	headingToTarget := math.Atan2(targetDelta.x, targetDelta.y)
	deltaHeading := normalizeAngle(headingToTarget - current.heading)

	// Turn in place until roughly facing the target.
	if math.Abs(deltaHeading) > math.Pi/6 {
		stepHeading(current, headingToTarget, dt)
		return
	}

	// Drive straight when the angular correction would be sub-millimeter.
	linearTol := math.Asin(0.001 * 2 / targetDist)
	if math.Abs(deltaHeading) <= linearTol {
		stepDelta := targetDelta.normalize().scale(robotMaxSpeed * dt.Seconds())
		if stepDelta.length() > targetDist {
			stepDelta = targetDelta
		}
		current.pos = current.pos.add(stepDelta)
		return
	}

	// Curved motion: compute wheel speeds for the arc that reaches the target.
	radius := targetDist / 2 / math.Sin(deltaHeading)
	k := robotWheelBaseRadius / 2 / radius
	ratioVrVl := (1 + k) / (1 - k)
	var vl, vr float64
	if math.Abs(ratioVrVl) >= 1 {
		vr = robotMaxSpeed
		vl = vr / ratioVrVl
	} else {
		vl = robotMaxSpeed
		vr = vl * ratioVrVl
	}

	linearSpeed := (vr + vl) / 2
	if linearSpeed < 0 {
		log.Panicf("sim-robot: unexpected negative linear speed: %f (radius=%f vr=%f vl=%f)", linearSpeed, radius, vr, vl)
	}
	if linearSpeed*dt.Seconds() > targetDist {
		linearSpeed = targetDist / dt.Seconds()
	}

	// sin/cos are swapped because heading 0 points along +Y, not +X.
	linearVel := v2{
		x: math.Sin(current.heading),
		y: math.Cos(current.heading),
	}.scale(linearSpeed)
	angularVel := linearSpeed / radius

	current.pos = current.pos.add(linearVel.scale(dt.Seconds()))
	current.heading += angularVel * dt.Seconds()
}

// stepHeading rotates current.heading toward targetHeading by at most one time step.
// Returns true when the heading is within tolerance.
func stepHeading(current *simPose, targetHeading float64, dt time.Duration) bool {
	delta := normalizeAngle(targetHeading - current.heading)
	if math.Abs(delta) < robotAngleTolerance {
		return true
	}
	step := min(robotMaxAngularSpeed*dt.Seconds(), math.Abs(delta))
	current.heading += math.Copysign(step, delta)
	return false
}

// normalizeAngle wraps an angle difference into [-pi, pi].
func normalizeAngle(delta float64) float64 {
	delta = math.Mod(delta, 2*math.Pi)
	if delta > math.Pi {
		delta -= 2 * math.Pi
	}
	if delta < -math.Pi {
		delta += 2 * math.Pi
	}
	return delta
}

// Package wscontroller implements fleetapi.FleetController over WebSocket using
// the robo-proto wire protocol. External robots or simulators open a persistent
// WebSocket connection, send HelloMessage to register, then exchange
// PoseUpdateMessage (robot→server) and SetTargetCommand / SetPoseCommand
// (server→robot) for fleet coordination.
//
// Usage:
//
//	ctrl := wscontroller.NewController()
//	mux.Handle("/api/robots/ws", ctrl)        // hook into existing HTTP server
//	go ctrl.Run(ctx, cmdChan, eventChan)      // start fleet controller loop
package wscontroller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"

	"checkerbots/apps/server/fleetapi"
	roboproto "checkerbots/packages/robo-proto"
)

var msgSeq atomic.Int64

func newMsgID() string {
	return fmt.Sprintf("msg-%d", msgSeq.Add(1))
}

// connID is an internal per-connection identifier, independent of robot_id.
// Using connID as the primary key means duplicate robot_id announcements are
// handled safely: each connection is tracked individually, and command routing
// to a robot_id uses last-connected-wins semantics.
type connID = uint64

var connSeq atomic.Uint64

type registration struct {
	id      connID
	robotID fleetapi.RobotID
	// sendCh is how Run pushes commands into the robot's write goroutine.
	sendCh chan<- fleetapi.Command
	// Signals when registration is complete (or fails) and commands are routed to this connection.
	ackCh chan<- registrationAck
}

type registrationAck struct {
	ok bool
}

type connEntry struct {
	robotID fleetapi.RobotID
	// sendCh is how Run pushes commands into the robot's write goroutine.
	sendCh chan<- fleetapi.Command
}

// Controller implements fleetapi.FleetController via WebSocket using robo-proto.
// It is also an http.Handler; mount it to accept robot connections.
type Controller struct {
	// registerCh delivers newly connected robots from ServeHTTP to Run.
	registerCh chan registration
	// unregisterCh delivers connIDs of disconnected robots to Run.
	unregisterCh chan connID
	cmdChan      <-chan fleetapi.Command
	eventChan    chan<- fleetapi.Event
}

// NewController returns a Controller ready to accept connections.
// Call Run in a goroutine before (or shortly after) mounting ServeHTTP.
func NewController(cmdChan <-chan fleetapi.Command, eventChan chan<- fleetapi.Event) *Controller {
	return &Controller{
		registerCh:   make(chan registration, 16),
		unregisterCh: make(chan connID, 16),
		cmdChan:      cmdChan,
		eventChan:    eventChan,
	}
}

// ServeHTTP upgrades the HTTP connection to WebSocket and manages one robot
// session. The robot must send HelloMessage as its very first message.
func (c *Controller) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// Accept any origin; add an origin allowlist if the endpoint is public-facing.
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Printf("ws-controller: accept: %v", err)
		return
	}
	defer conn.CloseNow()

	ctx := r.Context()

	// First message must be HelloMessage.
	_, raw, err := conn.Read(ctx)
	if err != nil {
		log.Printf("ws-controller: read hello from %s: %v", r.RemoteAddr, err)
		return
	}

	var env roboproto.Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		log.Printf("ws-controller: bad envelope from %s: %v", r.RemoteAddr, err)
		return
	}
	if env.Type != roboproto.MessageTypeHello {
		log.Printf("ws-controller: expected hello, got %q from %s", env.Type, r.RemoteAddr)
		conn.Close(websocket.StatusPolicyViolation, "first message must be hello")
		return
	}

	var hello roboproto.HelloMessage
	if err := json.Unmarshal(raw, &hello); err != nil {
		log.Printf("ws-controller: parse hello from %s: %v", r.RemoteAddr, err)
		return
	}

	id := connSeq.Add(1)
	robotID := fleetapi.RobotID(hello.RobotID)
	log.Printf("ws-controller: robot %q connected (conn=%d model=%s version=%s status=%s)",
		robotID, id, hello.RobotModel, hello.SoftwareVersion, hello.Status)

	// Per-connection command channel. Run writes to it; the write goroutine reads.
	cmdCh := make(chan fleetapi.Command, 8)
	registeredCh := make(chan registrationAck, 1)

	// Register with Run. Buffered channel ensures this never blocks ServeHTTP.
	c.registerCh <- registration{
		id: id, robotID: robotID,
		sendCh: cmdCh,
		ackCh:  registeredCh,
	}
	defer func() {
		// Use a short timeout so cleanup isn't silently lost if Run has exited,
		// but avoid blocking indefinitely on a cancelled request context.
		// TODO: Is this necessary? Is there a better way to handle this?
		select {
		case c.unregisterCh <- id:
		case <-time.After(5 * time.Second):
			log.Printf("ws-controller: timeout sending unregister for conn %d (robot %q)", id, robotID)
		}
	}()

	// Wait for ACK to start publishing events
	select {
	case ack := <-registeredCh:
		if !ack.ok {
			conn.Close(websocket.StatusPolicyViolation, "Robot ID already connected")
			return
		}
	case <-ctx.Done():
		return
	}

	c.eventChan <- fleetapi.RobotConnectedEvent{
		RobotID: robotID,
		CurrentPose: fleetapi.Pose{
			XMM:        hello.Pose.XMM,
			YMM:        hello.Pose.YMM,
			HeadingRad: hello.Pose.HeadingRad,
		},
		PoseValid: hello.PoseValid,
	}

	// Write goroutine: converts fleet commands into robo-proto wire messages.
	writeCtx, cancelWrite := context.WithCancel(ctx)
	defer cancelWrite()
	go c.writeLoop(writeCtx, conn, robotID, cmdCh)

	// Read loop: forwards robot events (pose updates, errors, …) to Run.
	if err := c.readLoop(ctx, conn, robotID); err != nil {
		log.Printf("ws-controller: robot %q disconnected: %v", robotID, err)
	}
}

// readLoop reads messages from the WebSocket and dispatches them.
// It returns when the connection closes or the context is cancelled.
func (c *Controller) readLoop(ctx context.Context, conn *websocket.Conn, robotID fleetapi.RobotID) error {
	for {
		_, raw, err := conn.Read(ctx)
		if err != nil {
			return err
		}

		var env roboproto.Envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			log.Printf("ws-controller: robot %q: bad envelope: %v", robotID, err)
			continue
		}

		switch env.Type {
		case roboproto.MessageTypePoseUpdate:
			var msg roboproto.PoseUpdateMessage
			if err := json.Unmarshal(raw, &msg); err != nil {
				log.Printf("ws-controller: robot %q: parse pose_update: %v", robotID, err)
				continue
			}
			event := fleetapi.PoseUpdateEvent{
				RobotID: robotID,
				CurrentPose: fleetapi.Pose{
					XMM:        msg.Pose.XMM,
					YMM:        msg.Pose.YMM,
					HeadingRad: msg.Pose.HeadingRad,
				},
			}
			select {
			case c.eventChan <- event:
			case <-ctx.Done():
				return ctx.Err()
			}
		case roboproto.MessageTypeHeartbeat:
			// Liveness only; no action needed.
		case roboproto.MessageTypeTelemetry:
			// Ignored for now.
		case roboproto.MessageTypeCommandResult:
			// TODO: correlate with dispatched commands for acknowledgement tracking.
		case roboproto.MessageTypeError:
			var msg roboproto.ErrorMessage
			if err := json.Unmarshal(raw, &msg); err == nil {
				log.Printf("ws-controller: robot %q error: %s: %s", robotID, msg.Code, msg.Message)
			}
		case roboproto.MessageTypePong:
			// Response to ping; no action needed.
		default:
			log.Printf("ws-controller: robot %q: unhandled message type %q", robotID, env.Type)
		}
	}
}

// writeLoop reads fleet commands from cmdCh and writes the corresponding
// robo-proto messages to the WebSocket. It exits when ctx is cancelled or
// cmdCh is closed.
func (c *Controller) writeLoop(ctx context.Context, conn *websocket.Conn, robotID fleetapi.RobotID, cmdCh <-chan fleetapi.Command) {
	for {
		select {
		case <-ctx.Done():
			return
		case cmd, ok := <-cmdCh:
			if !ok {
				// Closed cmdCh indicates server-initiated close
				// TODO: Use separate chan or event to signal server-initiated close with reason info
				conn.Close(websocket.StatusNormalClosure, "Closed by server")
				return
			}
			var data []byte
			var err error
			switch cmd := cmd.(type) {
			case fleetapi.SetTargetCommand:
				data, err = json.Marshal(roboproto.SetTargetCommand{
					CommandEnvelope: roboproto.CommandEnvelope{
						Envelope: roboproto.Envelope{
							ProtocolVersion: roboproto.ProtocolV0,
							Type:            roboproto.MessageTypeSetTarget,
							MessageID:       newMsgID(),
							RobotID:         string(cmd.RobotID),
							Timestamp:       time.Now().UTC(),
						},
						CommandID: newMsgID(),
					},
					Pose: roboproto.Pose{
						Pos:        roboproto.Pos{XMM: cmd.TargetPose.XMM, YMM: cmd.TargetPose.YMM},
						HeadingRad: cmd.TargetPose.HeadingRad,
					},
				})
			case fleetapi.SetPoseCommand:
				data, err = json.Marshal(roboproto.SetPoseCommand{
					CommandEnvelope: roboproto.CommandEnvelope{
						Envelope: roboproto.Envelope{
							ProtocolVersion: roboproto.ProtocolV0,
							Type:            roboproto.MessageTypeSetPose,
							MessageID:       newMsgID(),
							RobotID:         string(cmd.RobotID),
							Timestamp:       time.Now().UTC(),
						},
						CommandID: newMsgID(),
					},
					Pose: roboproto.Pose{
						Pos:        roboproto.Pos{XMM: cmd.CurrentPose.XMM, YMM: cmd.CurrentPose.YMM},
						HeadingRad: cmd.CurrentPose.HeadingRad,
					},
				})
			default:
				log.Printf("ws-controller: robot %q: unknown command type %T, dropping", robotID, cmd)
				continue
			}
			if err != nil {
				log.Printf("ws-controller: robot %q: marshal command: %v", robotID, err)
				continue
			}
			if err := conn.Write(ctx, websocket.MessageText, data); err != nil {
				log.Printf("ws-controller: robot %q: write error: %v", robotID, err)
				// TODO: Will a failed write close the socket or is an explicit unregister required?
				return
			}
		}
	}
}

// Run implements fleetapi.FleetController. Must be called in its own goroutine.
//
// It routes incoming fleet commands (cmdChan) to connected robot sessions and
// forwards robot events (pose updates, connect notifications) to the caller
// (eventChan).
func (c *Controller) Run(ctx context.Context) {
	// conns tracks all active connections keyed by their internal connID.
	conns := make(map[connID]connEntry)
	// robotRoute maps robot_id to the connID that currently handles it.
	// First-connected robot wins when the same robot ID reconnects.
	robotRoute := make(map[fleetapi.RobotID]connID)

	// Ensure all connections are finished before returning so that extra
	// unexpected events aren't published to eventChan
	defer func() {
		for _, conn := range conns {
			close(conn.sendCh)
		}

		for {
			select {
			case reg := <-c.registerCh:
				log.Printf("ws-controller: controller exiting, dropping new connection")
				reg.ackCh <- registrationAck{false}
			case id := <-c.unregisterCh:
				entry, ok := conns[id]
				if !ok {
					continue
				}
				close(entry.sendCh)
				delete(conns, id)
				// Only clear routing if this connID is still the active one for this robot;
				// a newer connection may have already taken over.
				if robotRoute[entry.robotID] == id {
					delete(robotRoute, entry.robotID)
				}
				log.Printf("ws-controller: unregistered robot %q (conn=%d, %d connected)", entry.robotID, id, len(conns))
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return

		case reg := <-c.registerCh:
			conns[reg.id] = connEntry{robotID: reg.robotID, sendCh: reg.sendCh}
			if _, exists := robotRoute[reg.robotID]; exists {
				log.Printf("ws-controller: robot %q already connected, dropping new connection", reg.robotID)
				reg.ackCh <- registrationAck{false}
				continue
			}
			robotRoute[reg.robotID] = reg.id
			log.Printf("ws-controller: registered robot %q (conn=%d, %d connected)", reg.robotID, reg.id, len(conns))
			reg.ackCh <- registrationAck{true}

		case id := <-c.unregisterCh:
			entry, ok := conns[id]
			if !ok {
				continue
			}
			close(entry.sendCh)
			delete(conns, id)
			// Only clear routing if this connID is still the active one for this robot;
			// a newer connection may have already taken over.
			if robotRoute[entry.robotID] == id {
				delete(robotRoute, entry.robotID)
			}
			log.Printf("ws-controller: unregistered robot %q (conn=%d, %d connected)", entry.robotID, id, len(conns))

		case cmd, ok := <-c.cmdChan:
			if !ok {
				return
			}
			var robotID fleetapi.RobotID
			switch cmd := cmd.(type) {
			case fleetapi.SetTargetCommand:
				robotID = cmd.RobotID
			case fleetapi.SetPoseCommand:
				robotID = cmd.RobotID
			default:
				log.Printf("ws-controller: unknown command type %T, dropping", cmd)
				continue
			}
			id, connected := robotRoute[robotID]
			if !connected {
				log.Printf("ws-controller: command for unconnected robot %q, dropping", robotID)
				continue
			}
			select {
			case conns[id].sendCh <- cmd:
			default:
				log.Printf("ws-controller: robot %q command channel full, dropping", robotID)
				// TODO: Close the connection immediately here?
			}
		}
	}
}

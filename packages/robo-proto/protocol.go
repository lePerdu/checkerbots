// Package protocol defines the shared robot protocol used by the coordination
// server, simulated robots, and future physical robot adapters.
//
// Protocol goals for v0:
//   - use the same message shapes for simulated and physical robots
//   - support registration, liveness, telemetry, pose updates, commands, and
//     command/result correlation
//   - keep transport details separate from the message schema
//
// Non-goals for v0:
//   - transport-specific framing rules
//   - authentication or encryption requirements
//   - advanced capability negotiation
//   - a full protocol version negotiation system
//   - standardized rich sensor schemas beyond the small generic telemetry shape
//
// Transport assumptions:
//   - transport-agnostic JSON message objects
//   - suitable for persistent bidirectional transports such as WebSocket or TCP
//   - robot_id is always included so messages remain self-describing in logs,
//     traces, and replayed event streams
//
// Encoding rules:
//   - timestamps use RFC3339 UTC and are represented here as time.Time for JSON
//     marshaling/unmarshaling
//   - IDs are opaque strings; UUIDs are recommended
//   - physical distances are millimeters
//   - heading angles are radians
//   - poses are expressed in a named coordinate frame when provided
//
// Behavioral notes:
//   - robots should send HelloMessage immediately after connecting
//   - HeartbeatMessage is intended to be lightweight and periodic
//   - TelemetryMessage and PoseUpdateMessage may be timer-driven, event-driven,
//     or both
//   - StopCommand is best-effort; inability to stop cleanly should still be
//     reported through CommandResultMessage or ErrorMessage
//   - SetPoseCommand updates believed pose and does not imply movement
//   - server policies like one-moving-robot-at-a-time belong in the coordinator,
//     not in the wire protocol
package roboproto

import "time"

type ProtocolVersion string

const ProtocolV0 ProtocolVersion = "robot-protocol.v0"

type MessageType string

type CommandType string

const (
	// Robot-to-server message types.
	MessageTypeHello         MessageType = "hello"
	MessageTypeHeartbeat     MessageType = "heartbeat"
	MessageTypeTelemetry     MessageType = "telemetry"
	MessageTypePoseUpdate    MessageType = "pose_update"
	MessageTypeCommandResult MessageType = "command_result"
	MessageTypeError         MessageType = "error"
	MessageTypePong          MessageType = "pong"
)

const (
	CommandTypeIdentify  CommandType = "identify"
	CommandTypeStop      CommandType = "stop"
	CommandTypeSetTarget CommandType = "set_target"
	CommandTypeSetPose   CommandType = "set_pose"
	CommandTypePing      CommandType = "ping"
)

const (
	// Server-to-robot command message types exposed as message types for envelope discrimination.
	MessageTypeIdentify   MessageType = MessageType(CommandTypeIdentify)
	MessageTypeStop       MessageType = MessageType(CommandTypeStop)
	MessageTypeMoveToPose MessageType = MessageType(CommandTypeSetTarget)
	MessageTypeSetPose    MessageType = MessageType(CommandTypeSetPose)
	MessageTypePing       MessageType = MessageType(CommandTypePing)
)

type PoseSource string

const (
	PoseSourceOdometry  PoseSource = "odometry"
	PoseSourceManual    PoseSource = "manual"
	PoseSourceSimulator PoseSource = "simulator"
)

type RobotStatus string

const (
	// These are the current recommended v0 status summary values.
	RobotStatusIdle          RobotStatus = "idle"
	RobotStatusMoving        RobotStatus = "moving"
	RobotStatusBlocked       RobotStatus = "blocked"
	RobotStatusError         RobotStatus = "error"
	RobotStatusManualControl RobotStatus = "manual_control"
)

type CommandStatus string

const (
	// CommandStatusAccepted means the robot has accepted the command for handling.
	CommandStatusAccepted CommandStatus = "accepted"
	// CommandStatusInProgress means execution has started but not yet completed.
	CommandStatusInProgress CommandStatus = "in_progress"
	// CommandStatusSucceeded means execution completed successfully.
	CommandStatusSucceeded CommandStatus = "succeeded"
	// CommandStatusFailed means execution completed unsuccessfully.
	CommandStatusFailed CommandStatus = "failed"
	// CommandStatusCancelled means execution was cancelled before successful completion.
	CommandStatusCancelled CommandStatus = "cancelled"
	// CommandStatusRejected means the robot refused the command before execution.
	CommandStatusRejected CommandStatus = "rejected"
)

// Envelope contains the fields required on all protocol messages.
//
// Required fields on all messages:
//   - protocol_version
//   - type
//   - message_id
//   - robot_id
//   - timestamp
//
// Timestamp uses time.Time's standard encoding/json behavior, which produces a
// quoted RFC3339-formatted timestamp string. When fractional seconds are
// present, Go emits enough digits to preserve the value, up to RFC3339Nano
// precision. The current offset is preserved, so callers should normalize to
// UTC before marshaling when they need strict v0 wire-format compliance such as
// "2026-08-08T12:00:05Z".
type Envelope struct {
	ProtocolVersion ProtocolVersion `json:"protocol_version"`
	Type            MessageType     `json:"type"`
	MessageID       string          `json:"message_id"`
	RobotID         string          `json:"robot_id"`
	Timestamp       time.Time       `json:"timestamp"`
}

// CommandEnvelope is embedded by all server-to-robot command messages.
//
// CommandID is required on every command so later CommandResultMessage and
// command-related ErrorMessage values can be correlated with dispatch and
// execution.
type CommandEnvelope struct {
	Envelope
	CommandID string `json:"command_id"`
}

// Pos is a transport-level position expressed in millimeters.
type Pos struct {
	XMM float64 `json:"x_mm"`
	YMM float64 `json:"y_mm"`
}

// Pose is a transport-level pose expressed in millimeters and radians.
type Pose struct {
	Pos
	HeadingRad float64 `json:"heading_rad"`
}

// TelemetrySnapshot is the intentionally small, generic v0 telemetry shape.
//
// Extras may contain adapter-specific flat or nested data.
type TelemetrySnapshot struct {
	BatteryPercent         float64        `json:"battery_percent"`
	LinearVelocityMMPerS   float64        `json:"linear_velocity_mm_per_s"`
	AngularVelocityRadPerS float64        `json:"angular_velocity_rad_per_s"`
	TemperatureC           float64        `json:"temperature_c"`
	Extras                 map[string]any `json:"extras,omitempty"`
}

// HelloMessage is sent when a robot first connects or reconnects.
//
// It registers presence, identifies robot kind/software version, and advertises
// the basic supported commands.
type HelloMessage struct {
	Envelope
	RobotModel      string      `json:"robot_model"`
	DisplayName     string      `json:"display_name"`
	SoftwareVersion string      `json:"software_version"`
	Status          RobotStatus `json:"status"`
}

// HeartbeatMessage is the periodic liveness message.
type HeartbeatMessage struct {
	Envelope
	Status        RobotStatus `json:"status"`
	LastCommandID string      `json:"last_command_id,omitempty"`
}

// TelemetryMessage carries periodic or change-driven telemetry snapshots.
type TelemetryMessage struct {
	Envelope
	TelemetrySnapshot
}

// PoseUpdateMessage reports the robot's latest pose estimate.
//
// When Confidence is present inside Pose, it is expected to be in the range
// [0.0, 1.0].
type PoseUpdateMessage struct {
	Envelope
	Pose       Pose       `json:"pose"`
	Source     PoseSource `json:"source"`
	Confidence float64    `json:"confidence"`
}

// CommandResultMessage reports command lifecycle progress and final outcomes.
//
// v0 allows multiple CommandResultMessage values for the same CommandID over
// time, such as accepted, then in_progress, then succeeded.
type CommandResultMessage struct {
	CommandEnvelope
	Status    CommandStatus  `json:"status"`
	Result    map[string]any `json:"result,omitempty"`
	ErrorCode string         `json:"error_code,omitempty"`
	Message   string         `json:"message,omitempty"`
}

// ErrorMessage reports protocol or runtime failures that are not adequately
// captured by telemetry alone.
//
// CommandID is optional in general, but should be set when the error is caused
// by a specific command.
type ErrorMessage struct {
	Envelope
	CommandID string         `json:"command_id,omitempty"`
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Details   map[string]any `json:"details,omitempty"`
}

// IdentifyCommand asks the robot to make itself easy for a human operator to identify.
type IdentifyCommand struct {
	CommandEnvelope
}

// StopCommand asks the robot to stop current motion as soon as possible.
type StopCommand struct {
	CommandEnvelope
}

// SetTargetCommand asks the robot to move to an explicit pose.
type SetTargetCommand struct {
	CommandEnvelope
	Pos Pos `json:"pos"`
}

// SetPoseCommand provides an authoritative pose fix, usually after manual
// setup, calibration, or recovery.
type SetPoseCommand struct {
	CommandEnvelope
	Pose   Pose   `json:"pose"`
	Source string `json:"source"`
}

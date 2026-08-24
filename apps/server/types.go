package main

import "time"

// CoordinateFrame names the frame a pose is expressed in.
type CoordinateFrame string

const (
	CoordinateFrameBoard CoordinateFrame = "board"
	CoordinateFrameWorld CoordinateFrame = "world"
)

// DefaultPoseFrame is the recommended default frame for MVP pose values.
const DefaultPoseFrame CoordinateFrame = CoordinateFrameBoard

// PoseSource describes where a pose estimate came from.
type PoseSource string

const (
	PoseSourceOdometry  PoseSource = "odometry"
	PoseSourceCamera    PoseSource = "camera"
	PoseSourceManual    PoseSource = "manual"
	PoseSourceSimulator PoseSource = "simulator"
	PoseSourceUnknown   PoseSource = "unknown"
)

// Pose is a board-relative or world-relative pose expressed in millimeters and radians.
//
// Source and Confidence are optional so the same type can be used both for
// measured robot poses and for command/calibration targets. When present,
// Confidence should be in the range [0.0, 1.0].
type Pose struct {
	XMM        float64         `json:"x_mm"`
	YMM        float64         `json:"y_mm"`
	ThetaRad   float64         `json:"theta_rad"`
	Frame      CoordinateFrame `json:"frame,omitempty"`
	Source     PoseSource      `json:"source,omitempty"`
	Confidence *float64        `json:"confidence,omitempty"`
}

// BoardCalibration captures the minimum data needed to map named board squares
// into physical space.
//
// Origin is the pose of the center of OriginSquare in the same frame used by
// board/world movement commands. A standard checkers board will usually use an
// origin square like a1 and square counts of 8x8, but the shape leaves room for
// other board sizes later.
type BoardCalibration struct {
	BoardID      string  `json:"board_id,omitempty"`
	OriginSquare string  `json:"origin_square"`
	Origin       Pose    `json:"origin"`
	SquareSizeMM float64 `json:"square_size_mm"`
	Columns      int     `json:"columns"`
	Rows         int     `json:"rows"`
}

// PieceAssignmentState tracks whether a logical piece currently has a robot.
type PieceAssignmentState string

const (
	PieceAssignmentStateAssigned   PieceAssignmentState = "assigned"
	PieceAssignmentStateUnassigned PieceAssignmentState = "unassigned"
)

// PieceAssignment links a logical game piece to the robot currently acting as it.
//
// RobotID is omitted when a piece is not assigned.
type PieceAssignment struct {
	GameID    string               `json:"game_id"`
	PieceID   string               `json:"piece_id"`
	RobotID   string               `json:"robot_id,omitempty"`
	State     PieceAssignmentState `json:"state"`
	UpdatedAt time.Time            `json:"updated_at"`
}

// RobotInfo holds the live state of a physical robot reported by external systems.
type RobotInfo struct {
	ID        string    `json:"id"`
	Pose      Pose      `json:"pose"`
	PieceID   string    `json:"piece_id"`
	UpdatedAt time.Time `json:"updated_at"`
}

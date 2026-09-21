package main

import fleetapi "checkerbots/apps/server/fleetapi"

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

type RobotID = fleetapi.RobotID

// Pose is a board-relative or world-relative pose expressed in millimeters and radians.
//
// Source and Confidence are optional so the same type can be used both for
// measured robot poses and for command/calibration targets. When present,
// Confidence should be in the range [0.0, 1.0].
type Pose struct {
	XMM float64 `json:"x_mm"`
	YMM float64 `json:"y_mm"`
	// HeadingRad is the robot's facing direction in radians.
	// 0 points from the black side toward the red side (+Y in world coordinates).
	// Angles increase clockwise when viewed from above.
	HeadingRad float64 `json:"heading_rad"`
}

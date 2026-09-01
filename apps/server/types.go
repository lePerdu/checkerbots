package main

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

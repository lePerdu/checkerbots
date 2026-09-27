package fleetapi

type RobotID string

type Pose struct {
	XMM float64
	YMM float64
	// HeadingRad is the robot's facing direction in radians.
	// 0 points from the black side toward the red side (+Y in world coordinates).
	// Angles increase clockwise when viewed from above.
	HeadingRad float64
}

// TODO: Add interface to better typecheck?
type Command any

type SetPoseCommand struct {
	RobotID     RobotID
	CurrentPose Pose
}

type SetTargetCommand struct {
	RobotID    RobotID
	TargetPose Pose
}

// TODO: Add interface to better typecheck?
type Event any

type RobotConnectedEvent struct {
	RobotID     RobotID
	CurrentPose Pose
	PoseValid   bool
}

type PoseUpdateEvent struct {
	RobotID     RobotID
	CurrentPose Pose
}

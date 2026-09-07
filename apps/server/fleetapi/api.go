package fleetapi

import (
	"context"
)

type RobotID string

type Pose struct {
	XMM float64
	YMM float64
}

type RobotCommand struct {
	RobotID    RobotID
	TargetPose Pose
}

type RobotUpdate struct {
	RobotID     RobotID
	CurrentPose Pose
}

type FleetController interface {
	Run(ctx context.Context, cmdChan <-chan RobotCommand, updateChan chan<- RobotUpdate)
}

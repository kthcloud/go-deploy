package models

import (
	"go-deploy/pkg/imp/harbor/sdk/v2.0/models"
	"strings"
	"time"
)

type RobotPublic struct {
	ID         int       `bson:"id"`
	Name       string    `bson:"name"`
	HarborName string    `bson:"harborName"`
	Disable    bool      `bson:"disable"`
	Secret     string    `bson:"secret" `
	CreatedAt  time.Time `bson:"createdAt"`
}

func (r *RobotPublic) Created() bool {
	return r.ID != 0
}

func (r *RobotPublic) IsPlaceholder() bool {
	return false
}

// CreateRobotPublicFromGet converts a modelv2.Robot to a RobotPublic.
func CreateRobotPublicFromGet(robot *models.Robot) *RobotPublic {
	return &RobotPublic{
		ID:         int(robot.ID),
		Name:       extractRobotRealName(robot.Name),
		HarborName: robot.Name,
		Disable:    robot.Disable,
		Secret:     robot.Description,
		CreatedAt:  time.Time(robot.CreationTime),
	}
}

// extractRobotRealName extracts the real name of a robot from the name returned by the API.
// This is because the API returns the name as <project name>$<robot name>+<project name>
func extractRobotRealName(name string) string {
	robotAndProject := strings.Split(name, "$")
	if len(robotAndProject) != 2 {
		return ""
	}

	onlyRobot := strings.Split(robotAndProject[1], "+")
	if len(onlyRobot) != 2 {
		return ""
	}

	return onlyRobot[1]
}

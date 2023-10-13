package models

import (
	"fmt"
	modelv2 "github.com/mittwald/goharbor-client/v5/apiv2/model"
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

func CreateRobotPublicFromGet(robot *modelv2.Robot, project *modelv2.Project) *RobotPublic {
	return &RobotPublic{
		ID:         int(robot.ID),
		Name:       extractRobotRealName(robot.Name),
		HarborName: robot.Name,
		Disable:    robot.Disable,
		Secret:     robot.Description,
		CreatedAt:  time.Time(robot.CreationTime),
	}
}

func getRobotFullName(projectName, name string) string {
	return fmt.Sprintf("robot$%s", getRobotName(projectName, name))
}

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

func getRobotName(projectName, name string) string {
	return fmt.Sprintf("%s+%s", projectName, name)
}

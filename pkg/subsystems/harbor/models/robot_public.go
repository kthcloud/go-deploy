package models

import (
	"fmt"
	modelv2 "github.com/mittwald/goharbor-client/v5/apiv2/model"
	"strings"
)

type RobotPublic struct {
	ID          int    `json:"id" bson:"id"`
	Name        string `json:"name" bson:"name"`
	HarborName  string `json:"harborName" bson:"harborName"`
	ProjectID   int    `json:"projectId" bson:"projectId"`
	ProjectName string `json:"projectName" bson:"projectName"`
	Description string `json:"description" bson:"description"`
	Disable     bool   `json:"disable" bson:"disable"`
	Secret      string `json:"secret" bson:"secret" `
}

func CreateRobotUpdateFromPublic(public *RobotPublic) *modelv2.Robot {
	return &modelv2.Robot{
		ID:          int64(public.ID),
		Name:        getRobotFullName(public.ProjectName, public.Name),
		Level:       "project",
		Description: public.Description,
		Disable:     public.Disable,
		Editable:    true,
		ExpiresAt:   -1,
		Permissions: getPermissions(public.Name),
	}
}

func CreateRobotPublicFromGet(robot *modelv2.Robot, project *modelv2.Project) *RobotPublic {
	return &RobotPublic{
		ID:          int(robot.ID),
		Name:        extractRobotRealName(robot.Name),
		HarborName:  robot.Name,
		ProjectName: project.Name,
		ProjectID:   int(project.ProjectID),
		Description: robot.Description,
		Disable:     robot.Disable,
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

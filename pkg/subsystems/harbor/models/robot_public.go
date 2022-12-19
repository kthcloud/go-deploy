package models

import modelv2 "github.com/mittwald/goharbor-client/v5/apiv2/model"

type RobotPublic struct {
	ID          int    `json:"id" bson:"id"`
	Name        string `json:"name" bson:"name"`
	ProjectID   int    `json:"projectId" bson:"projectId"`
	ProjectName string `json:"projectName" bson:"projectName"`
	Description string `json:"description" bson:"description"`
	Disable     bool   `json:"disable" bson:"disable"`
	Secret      string `json:"secret" bson:"secret" `
}

func CreateRobotParamsFromPublic(public *RobotPublic) *modelv2.Robot {
	return &modelv2.Robot{
		Description: public.Description,
		Disable:     public.Disable,
		Editable:    true,
		ExpiresAt:   -1,
		ID:          int64(public.ID),
		Level:       "project",
		Name:        public.Name,
		Permissions: getPermissions(public.Name),
	}
}

func CreateRobotPublicFromGet(robot *modelv2.Robot, project *modelv2.Project) *RobotPublic {
	return &RobotPublic{
		ID:          int(robot.ID),
		Name:        robot.Name,
		ProjectName: project.Name,
		ProjectID:   int(project.ProjectID),
		Description: robot.Description,
		Disable:     robot.Disable,
	}
}

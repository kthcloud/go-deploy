package models

import (
	"github.com/mittwald/goharbor-client/v5/apiv2/model"
)

type RobotCreated struct {
	ID     int    `json:"id"`
	Secret string `json:"secret"`
}

func CreateRobotCreateBody(public *RobotPublic, projectName string) *model.RobotCreate {
	return &model.RobotCreate{
		// These are updated in assertCorrectRobotSecret
		Description: "",
		Secret:      "",

		Disable:  public.Disable,
		Duration: -1,
		Level:    "project",
		Name:     public.Name,
		Permissions: []*model.RobotPermission{
			{
				Access: []*model.Access{
					{
						Action:   "list",
						Resource: "repository",
					},
					{
						Action:   "pull",
						Resource: "repository",
					},
					{
						Action:   "push",
						Resource: "repository",
					},
					{
						Action:   "create",
						Resource: "tag",
					},
				},
				Kind:      "project",
				Namespace: projectName,
			},
		},
	}
}

func CreateRobotUpdateBody(public *RobotPublic, projectName string) *model.Robot {
	return &model.Robot{
		// These are updated in assertCorrectRobotSecret
		Description: "",
		Secret:      "",

		ID:        int64(public.ID),
		Disable:   public.Disable,
		Name:      public.HarborName,
		ExpiresAt: -1,
		Duration:  -1,
		Level:     "project",
		Permissions: []*model.RobotPermission{
			{
				Access: []*model.Access{
					{
						Action:   "list",
						Resource: "repository",
					},
					{
						Action:   "pull",
						Resource: "repository",
					},
					{
						Action:   "push",
						Resource: "repository",
					},
					{
						Action:   "create",
						Resource: "tag",
					},
				},
				Kind:      "project",
				Namespace: projectName,
			},
		},
	}
}

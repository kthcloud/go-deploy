package models

import (
	"github.com/mittwald/goharbor-client/v5/apiv2/model"
)

func CreateRobotCreateReq(projectName, name string) model.RobotCreate {
	return model.RobotCreate{
		Description: "Automatically created",
		Disable:     false,
		Duration:    -1,
		Level:       "project",
		Name:        name,
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

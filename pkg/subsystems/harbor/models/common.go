package models

import "github.com/mittwald/goharbor-client/v5/apiv2/model"

func getPermissions(projectName string) []*model.RobotPermission {
	return []*model.RobotPermission{
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
	}
}

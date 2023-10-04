package models

import (
	"github.com/mittwald/goharbor-client/v5/apiv2/model"
)

func CreateRobotCreateBody(public *RobotPublic) model.RobotCreate {
	return model.RobotCreate{
		Description: "",
		Disable:     public.Disable,
		Duration:    -1,
		Level:       "project",
		Name:        public.Name,
		Secret:      public.Secret,
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
				Namespace: public.ProjectName,
			},
		},
	}
}

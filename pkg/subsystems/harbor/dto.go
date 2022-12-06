package harbor

import (
	"github.com/mittwald/goharbor-client/v5/apiv2/model"
	"go-deploy/utils/subsystemutils"
)

func createProjectRequestBody(projectName string) model.ProjectReq {
	return model.ProjectReq{
		ProjectName:  projectName,
		StorageLimit: int64Ptr(0),
	}
}

func createRobotRequestBody(name string) model.RobotCreate {
	prefixedName := subsystemutils.GetPrefixedName(name)
	return model.RobotCreate{
		Description: "Robot Account for deploy project",
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
				Namespace: prefixedName,
			},
		},
	}
}

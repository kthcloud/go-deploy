package models

import (
	"github.com/kthcloud/go-deploy/pkg/imp/harbor/sdk/v2.0/models"
	"strconv"
)

// CreateProjectCreateBody creates a body used for create a project in the Harbor API.
func CreateProjectCreateBody(public *ProjectPublic) models.ProjectReq {
	return models.ProjectReq{
		ProjectName:  public.Name,
		StorageLimit: new(int64(-1)),
		Metadata: &models.ProjectMetadata{
			Public: strconv.FormatBool(public.Public),
		},
	}
}

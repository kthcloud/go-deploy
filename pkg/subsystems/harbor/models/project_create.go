package models

import (
	"github.com/mittwald/goharbor-client/v5/apiv2/model"
	"strconv"
)

func int64Ptr(i int64) *int64 { return &i }

// CreateProjectCreateBody creates a body used for create a project in the Harbor API.
func CreateProjectCreateBody(public *ProjectPublic) model.ProjectReq {
	return model.ProjectReq{
		ProjectName:  public.Name,
		StorageLimit: int64Ptr(0),
		Metadata: &model.ProjectMetadata{
			Public: strconv.FormatBool(public.Public),
		},
	}
}

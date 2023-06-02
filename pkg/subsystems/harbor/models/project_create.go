package models

import "github.com/mittwald/goharbor-client/v5/apiv2/model"

func int64Ptr(i int64) *int64 { return &i }

func CreateProjectCreateBody(public *ProjectPublic) model.ProjectReq {
	return model.ProjectReq{
		ProjectName:  public.Name,
		StorageLimit: int64Ptr(0),
		Metadata: &model.ProjectMetadata{
			Public: "true",
		},
	}
}

package models

import "github.com/mittwald/goharbor-client/v5/apiv2/model"

func int64Ptr(i int64) *int64 { return &i }

func CreateProjectCreateReq(projectName string) model.ProjectReq {
	return model.ProjectReq{
		ProjectName:  projectName,
		StorageLimit: int64Ptr(0),
	}
}

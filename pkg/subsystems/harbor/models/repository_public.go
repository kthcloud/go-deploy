package models

import modelv2 "github.com/mittwald/goharbor-client/v5/apiv2/model"

type RepositoryPublic struct {
	ID          int          `json:"ID" bson:"ID"`
	Name        string       `json:"name" bson:"name"`
	ProjectID   int          `json:"projectId" bson:"projectId"`
	ProjectName string       `json:"projectName" bson:"projectName"`
	Seeded      bool         `json:"seeded" bson:"seeded"`
	Placeholder *PlaceHolder `json:"placeholder" bson:"placeholder"`
}

func CreateRepositoryPublicFromGet(repository *modelv2.Repository, project *modelv2.Project) *RepositoryPublic {
	return &RepositoryPublic{
		ID:          int(repository.ID),
		Name:        repository.Name,
		ProjectID:   int(project.ProjectID),
		ProjectName: project.Name,
		Seeded:      repository.ArtifactCount > 0,
	}
}

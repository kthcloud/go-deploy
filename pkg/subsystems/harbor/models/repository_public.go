package models

import (
	modelv2 "github.com/mittwald/goharbor-client/v5/apiv2/model"
	"time"
)

type RepositoryPublic struct {
	ID          int          `bson:"id"`
	Name        string       `bson:"name"`
	Seeded      bool         `bson:"seeded"`
	Placeholder *PlaceHolder `bson:"placeholder"`
	CreatedAt   time.Time    `bson:"createdAt"`
}

func (r *RepositoryPublic) Created() bool {
	return r.ID != 0
}

func (r *RepositoryPublic) IsPlaceholder() bool {
	return false
}

func CreateRepositoryPublicFromGet(repository *modelv2.Repository, project *modelv2.Project) *RepositoryPublic {
	var createdAt time.Time
	if repository.CreationTime != nil {
		createdAt = time.Time(*repository.CreationTime)
	}

	return &RepositoryPublic{
		ID:        int(repository.ID),
		Name:      repository.Name,
		Seeded:    repository.ArtifactCount > 0,
		CreatedAt: createdAt,
	}
}

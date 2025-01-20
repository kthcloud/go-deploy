package models

import (
	"github.com/kthcloud/go-deploy/pkg/imp/harbor/sdk/v2.0/models"
	"strings"
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

// CreateRepositoryPublicFromGet converts a modelv2.Repository to a RepositoryPublic.
func CreateRepositoryPublicFromGet(repository *models.Repository) *RepositoryPublic {
	var createdAt time.Time
	if repository.CreationTime != nil {
		createdAt = time.Time(*repository.CreationTime)
	}

	return &RepositoryPublic{
		ID:        int(repository.ID),
		Name:      extractRepositoryName(repository.Name),
		Seeded:    repository.ArtifactCount > 0,
		CreatedAt: createdAt,
	}
}

// For some reason, the name is returned as name=<project name>/<repo name>
// even though it is used as only <repo name> in the api
//
// Lovely Harbor API!
// :)
func extractRepositoryName(name string) string {
	split := strings.Split(name, "/")
	if len(split) == 2 {
		return split[1]
	}

	return name
}

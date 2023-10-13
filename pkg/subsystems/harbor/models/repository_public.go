package models

import (
	modelv2 "github.com/mittwald/goharbor-client/v5/apiv2/model"
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

func CreateRepositoryPublicFromGet(repository *modelv2.Repository) *RepositoryPublic {
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

// for some reason, the name is returned as name=<project name>/<repo name>
// even though it is used as only <repo name> in the api
// lovely harbor api :)
func extractRepositoryName(name string) string {
	split := strings.Split(name, "/")
	if len(split) == 2 {
		return split[1]
	}

	return name
}

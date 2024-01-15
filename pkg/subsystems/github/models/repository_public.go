package models

import "github.com/google/go-github/github"

type RepositoryPublic struct {
	ID            int64  `bson:"id"`
	Name          string `bson:"name"`
	Owner         string `bson:"owner"`
	CloneURL      string `bson:"cloneUrl"`
	DefaultBranch string `bson:"defaultBranch"`
}

// CreateRepositoryPublicFromRead converts a github.Repository to a RepositoryPublic.
func CreateRepositoryPublicFromRead(repo *github.Repository) *RepositoryPublic {
	return &RepositoryPublic{
		ID:            *repo.ID,
		Name:          *repo.Name,
		Owner:         *repo.Owner.Login,
		CloneURL:      *repo.CloneURL,
		DefaultBranch: *repo.DefaultBranch,
	}
}

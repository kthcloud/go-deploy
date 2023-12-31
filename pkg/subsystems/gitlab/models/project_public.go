package models

import "github.com/xanzy/go-gitlab"

type ProjectPublic struct {
	ID        int    `bson:"id"`
	Name      string `bson:"name"`
	ImportURL string `bson:"importUrl"`
}

func CreateProjectPublicFromGet(p *gitlab.Project) *ProjectPublic {
	return &ProjectPublic{
		ID:        p.ID,
		Name:      p.Name,
		ImportURL: p.ImportURL,
	}
}

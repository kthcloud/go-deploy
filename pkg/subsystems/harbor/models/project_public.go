package models

import (
	modelv2 "github.com/mittwald/goharbor-client/v5/apiv2/model"
	"time"
)

type ProjectPublic struct {
	ID        int       `bson:"id"`
	Name      string    `bson:"name"`
	CreatedAt time.Time `bson:"createdAt"`
}

func (p *ProjectPublic) Created() bool {
	return p.ID != 0
}

func CreateProjectUpdateParamsFromPublic(public *ProjectPublic) *modelv2.Project {
	return &modelv2.Project{
		Name: public.Name,
	}
}

func CreateProjectPublicFromGet(project *modelv2.Project) *ProjectPublic {
	return &ProjectPublic{
		ID:        int(project.ProjectID),
		Name:      project.Name,
		CreatedAt: time.Time(project.CreationTime),
	}
}

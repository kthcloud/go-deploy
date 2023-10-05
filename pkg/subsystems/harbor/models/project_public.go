package models

import (
	modelv2 "github.com/mittwald/goharbor-client/v5/apiv2/model"
	"strconv"
	"time"
)

type ProjectPublic struct {
	ID        int       `bson:"id"`
	Name      string    `bson:"name"`
	Public    bool      `bson:"public"`
	CreatedAt time.Time `bson:"createdAt"`
}

func (p *ProjectPublic) Created() bool {
	return p.ID != 0
}

func CreateProjectUpdateParamsFromPublic(public *ProjectPublic) *modelv2.Project {
	return &modelv2.Project{
		Name: public.Name,
		Metadata: &modelv2.ProjectMetadata{
			Public: strconv.FormatBool(public.Public),
		},
	}
}

func CreateProjectPublicFromGet(project *modelv2.Project) *ProjectPublic {
	return &ProjectPublic{
		ID:        int(project.ProjectID),
		Name:      project.Name,
		Public:    project.Metadata.Public == "true",
		CreatedAt: time.Time(project.CreationTime),
	}
}

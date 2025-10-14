package models

import (
	"github.com/kthcloud/go-deploy/pkg/imp/harbor/sdk/v2.0/models"
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

func (p *ProjectPublic) IsPlaceholder() bool {
	return false
}

// CreateProjectUpdateParamsFromPublic creates a body used for update a project in the Harbor API.
func CreateProjectUpdateParamsFromPublic(public *ProjectPublic) *models.ProjectReq {
	return &models.ProjectReq{
		Metadata: &models.ProjectMetadata{
			Public: strconv.FormatBool(public.Public),
		},
	}
}

// CreateProjectPublicFromGet converts a modelv2.Project to a ProjectPublic.
func CreateProjectPublicFromGet(project *models.Project) *ProjectPublic {
	return &ProjectPublic{
		ID:        int(project.ProjectID),
		Name:      project.Name,
		Public:    project.Metadata.Public == "true",
		CreatedAt: time.Time(project.CreationTime),
	}
}

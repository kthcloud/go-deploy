package models

import modelv2 "github.com/mittwald/goharbor-client/v5/apiv2/model"

type ProjectPublic struct {
	ID     int    `json:"id,omitempty" bson:"id"`
	Name   string `json:"name,omitempty" bson:"name"`
	Public bool   `json:"public,omitempty" bson:"public,omitempty"`
}

func (p *ProjectPublic) Created() bool {
	return p.ID != 0
}

func CreateProjectUpdateParamsFromPublic(public *ProjectPublic) *modelv2.Project {
	return &modelv2.Project{
		Metadata: &modelv2.ProjectMetadata{
			Public: boolToString(public.Public),
		},
		Name: public.Name,
	}
}

func CreateProjectPublicFromGet(project *modelv2.Project) *ProjectPublic {
	return &ProjectPublic{
		ID:     int(project.ProjectID),
		Name:   project.Name,
		Public: stringToBool(project.Metadata.Public),
	}
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func stringToBool(s string) bool {
	return s == "true"
}

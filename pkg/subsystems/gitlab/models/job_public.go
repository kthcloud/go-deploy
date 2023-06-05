package models

import (
	"github.com/xanzy/go-gitlab"
	"time"
)

type JobPublic struct {
	ID        int       `json:"id"`
	ProjectID int       `json:"projectId"`
	Status    string    `json:"status"`
	Stage     string    `json:"stage"`
	CreatedAt time.Time `json:"createdAt"`
}

func CreateJobPublicFromGet(job *gitlab.Job) *JobPublic {
	var createdAt time.Time
	if job.CreatedAt != nil {
		createdAt = *job.CreatedAt
	}

	var projectID int
	if job.Project != nil {
		projectID = job.Project.ID
	}

	return &JobPublic{
		ID:        job.ID,
		ProjectID: projectID,
		Status:    job.Status,
		Stage:     job.Stage,
		CreatedAt: createdAt,
	}
}

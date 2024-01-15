package models

import (
	"github.com/xanzy/go-gitlab"
	"time"
)

type JobPublic struct {
	ID        int       `bson:"id"`
	ProjectID int       `bson:"projectId"`
	Status    string    `bson:"status"`
	Stage     string    `bson:"stage"`
	CreatedAt time.Time `bson:"createdAt"`
}

// CreateJobPublicFromGet converts a gitlab.Job to a JobPublic.
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

package models

import "github.com/xanzy/go-gitlab"

type JobPublic struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
	Stage  string `json:"stage"`
}

func CreateJobPublicFromGet(job *gitlab.Job) *JobPublic {
	return &JobPublic{
		ID:     job.ID,
		Status: job.Status,
		Stage:  job.Stage,
	}
}

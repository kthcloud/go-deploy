package models

import "time"

type LogLine struct {
	DeploymentName string
	PodName        string
	Line           string
	CreatedAt      time.Time
}

type PodDeleted struct {
	DeploymentName string
	PodName        string
}

package models

import "time"

type LogLine struct {
	Line      string
	PodNumber int
	CreatedAt time.Time
}

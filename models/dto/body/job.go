package body

import "time"

type JobRead struct {
	ID         string     `json:"id"`
	UserID     string     `json:"userId"`
	Type       string     `json:"type"`
	Status     string     `json:"status"`
	LastError  *string    `json:"lastError,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	LastRunAt  *time.Time `json:"lastRunAt,omitempty"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
	RunAfter   time.Time  `json:"runAfter,omitempty"`
}

type JobUpdate struct {
	Status *string `json:"status" binding:"omitempty,oneof=pending running failed terminated finished"`
}

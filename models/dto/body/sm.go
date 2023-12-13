package body

import "time"

type SmDeleted struct {
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}

type SmRead struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"ownerId"`
	CreatedAt time.Time `json:"createdAt"`
	Zone      string    `json:"zone"`
	URL       *string   `json:"url,omitempty"`
}

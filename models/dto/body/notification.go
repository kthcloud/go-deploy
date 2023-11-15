package body

import "time"

type NotificationRead struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Content   map[string]interface{} `json:"content"`
	CreatedAt time.Time              `json:"createdAt"`
	ReadAt    *time.Time             `json:"readAt,omitempty"`
}

type NotificationUpdate struct {
	Read bool `json:"read"`
}

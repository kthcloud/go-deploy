package body

import "time"

type NotificationRead struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"userId"`
	Type        string                 `json:"type"`
	Content     map[string]interface{} `json:"content"`
	CreatedAt   time.Time              `json:"createdAt"`
	ReadAt      *time.Time             `json:"readAt,omitempty"`
	CompletedAt *time.Time             `json:"completedAt,omitempty"`
}

type NotificationUpdate struct {
	Read bool `json:"read"`
}

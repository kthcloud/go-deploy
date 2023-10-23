package body

import "time"

type NotificationRead struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Content map[string]interface{} `json:"content"`
	ReadAt  *time.Time             `json:"readAt,omitempty"`
}

type NotificationUpdate struct {
	ReadAt *time.Time `json:"readAt"`
}

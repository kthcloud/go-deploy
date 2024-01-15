package notification

import (
	"go-deploy/models/dto/body"
	"time"
)

// ToDTO converts a Notification to a body.NotificationRead DTO.
func (notification *Notification) ToDTO() body.NotificationRead {
	var readAt *time.Time
	if !notification.ReadAt.IsZero() {
		readAt = &notification.ReadAt
	}

	var completedAt *time.Time
	if !notification.CompletedAt.IsZero() {
		completedAt = &notification.CompletedAt
	}

	return body.NotificationRead{
		ID:          notification.ID,
		UserID:      notification.UserID,
		Type:        notification.Type,
		Content:     notification.Content,
		CreatedAt:   notification.CreatedAt,
		ReadAt:      readAt,
		CompletedAt: completedAt,
	}
}

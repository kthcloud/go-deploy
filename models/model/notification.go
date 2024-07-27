package model

import (
	"go-deploy/dto/v2/body"
	"time"
)

const (
	// NotificationTeamInvite is used for team invite notifications.
	NotificationTeamInvite = "teamInvite"
	// NotificationResourceTransfer is used for resource migration notifications.
	NotificationResourceTransfer = "resourceTransfer"
)

type Notification struct {
	ID      string                 `bson:"id"`
	UserID  string                 `bson:"userId"`
	Type    string                 `bson:"type"`
	Content map[string]interface{} `bson:"content"`

	CreatedAt   time.Time `bson:"createdAt"`
	ReadAt      time.Time `bson:"readAt,omitempty"`
	ToastedAt   time.Time `bson:"toastedAt,omitempty"`
	CompletedAt time.Time `bson:"completedAt,omitempty"`
	DeletedAt   time.Time `bson:"deletedAt,omitempty"`
}

type NotificationCreateParams struct {
	Type    string                 `json:"type"`
	Content map[string]interface{} `json:"content"`
}

// ToDTO converts a Notification to a body.NotificationRead DTO.
func (notification *Notification) ToDTO() body.NotificationRead {
	var readAt *time.Time
	if !notification.ReadAt.IsZero() {
		readAt = &notification.ReadAt
	}

	var toastedAt *time.Time
	if !notification.ToastedAt.IsZero() {
		toastedAt = &notification.ToastedAt
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
		ToastedAt:   toastedAt,
		CompletedAt: completedAt,
	}
}

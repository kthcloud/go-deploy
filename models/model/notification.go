package model

import (
	body "github.com/kthcloud/go-deploy/api/openapi/v2"
	"time"
)

const (
	// NotificationTeamInvite is used for team invite notifications.
	NotificationTeamInvite = "teamInvite"
	// NotificationResourceTransfer is used for resource migration notifications.
	NotificationResourceTransfer = "resourceTransfer"
)

type Notification struct {
	ID      string         `bson:"id"`
	UserID  string         `bson:"userId"`
	Type    string         `bson:"type"`
	Content map[string]any `bson:"content"`

	CreatedAt   time.Time `bson:"createdAt"`
	ReadAt      time.Time `bson:"readAt,omitempty"`
	ToastedAt   time.Time `bson:"toastedAt,omitempty"`
	CompletedAt time.Time `bson:"completedAt,omitempty"`
	DeletedAt   time.Time `bson:"deletedAt,omitempty"`
}

type NotificationCreateParams struct {
	Type    string         `json:"type"`
	Content map[string]any `json:"content"`
}

// ToDTO converts a Notification to a body.NotificationRead DTO.
func (notification *Notification) ToDTO() body.NotificationRead {
	return dtoConverter.NotificationToDTO(notification)
}

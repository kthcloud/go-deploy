package model

import (
	"go-deploy/dto/v1/body"
	"time"
)

const (
	// NotificationTeamInvite is used for team invite notifications.
	NotificationTeamInvite = "teamInvite"
	// NotificationDeploymentTransfer is used for deployment transfer notifications.
	NotificationDeploymentTransfer = "deploymentTransfer"
	// NotificationVmTransfer is used for vm transfer notifications.
	NotificationVmTransfer = "vmTransfer"
)

type Notification struct {
	ID      string                 `bson:"id"`
	UserID  string                 `bson:"userId"`
	Type    string                 `bson:"type"`
	Content map[string]interface{} `bson:"content"`

	CreatedAt   time.Time `bson:"createdAt"`
	ReadAt      time.Time `bson:"readAt,omitempty"`
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

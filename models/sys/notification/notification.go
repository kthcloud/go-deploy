package notification

import (
	"time"
)

const (
	TypeTeamInvite         = "teamInvite"
	TypeDeploymentTransfer = "deploymentTransfer"
	TypeVmTransfer         = "vmTransfer"
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

type CreateParams struct {
	Type    string                 `json:"type"`
	Content map[string]interface{} `json:"content"`
}

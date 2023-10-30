package notification

import (
	"go-deploy/models/dto/body"
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

	ReadAt    *time.Time `bson:"readAt,omitempty"`
	DeletedAt *time.Time `bson:"deletedAt,omitempty"`
}

type CreateParams struct {
	Type    string                 `json:"type"`
	Content map[string]interface{} `json:"content"`
}

type UpdateParams struct {
	ReadAt *time.Time `json:"readAt"`
}

func (u *UpdateParams) FromDTO(dto *body.NotificationUpdate) {
	u.ReadAt = dto.ReadAt
}

package model

import (
	"go-deploy/dto/v2/body"
	"time"
)

const (
	ResourceMigrationTypeUpdateOwner = "updateOwner"

	ResourceMigrationResourceTypeVM         = "vm"
	ResourceMigrationResourceTypeDeployment = "deployment"

	ResourceMigrationStatusPending  = "pending"
	ResourceMigrationStatusAccepted = "accepted"
)

type ResourceMigration struct {
	ID         string `bson:"id"`
	ResourceID string `bson:"resourceId"`
	UserID     string `bson:"userId"`

	Type         string `bson:"type"`
	ResourceType string `bson:"resourceType"`
	Status       string `bson:"status"`

	Code *string `bson:"code"`

	CreatedAt time.Time  `bson:"createdAt"`
	DeletedAt *time.Time `bson:"deletedAt,omitempty"`

	UpdateOwner *ResourceMigrationUpdateOwner `bson:"updateOwner,omitempty"`
}

type ResourceMigrationUpdateOwner struct {
	NewOwnerID string
	OldOwnerID string
}

type ResourceMigrationUpdateOwnerParams struct {
	NewOwnerID string
	OldOwnerID string
}

// ToDTO returns the resource migration to a body.ResourceMigrationDTO
func (r *ResourceMigration) ToDTO() body.ResourceMigrationRead {
	return body.ResourceMigrationRead{
		ID:         r.ID,
		ResourceID: r.ResourceID,
		UserID:     r.UserID,

		Type:         r.Type,
		ResourceType: r.ResourceType,
		Status:       r.Status,

		UpdateOwner: &struct {
			OwnerID string `json:"ownerId"`
		}{
			OwnerID: r.UpdateOwner.NewOwnerID,
		},

		CreatedAt: r.CreatedAt,
		DeletedAt: r.DeletedAt,
	}
}

package model

import (
	"go-deploy/dto/v1/body"
	"time"
)

const (
	ResourceMigrationTypeUpdateOwner = "updateOwner"

	ResourceMigrationResourceTypeVM         = "vm"
	ResourceMigrationResourceTypeDeployment = "deployment"

	ResourceMigrationStatusPending  = "pending"
	ResourceMigrationStatusReady    = "ready"
	ResourceMigrationStatusRunning  = "running"
	ResourceMigrationStatusComplete = "complete"
)

type ResourceMigration struct {
	ID         string `bson:"id"`
	ResourceID string `bson:"resourceId"`
	UserID     string `bson:"userId"`

	Type         string `bson:"type"`
	ResourceType string `bson:"resourceType"`
	Status       string `bson:"status"`

	Token *string `bson:"token"`

	CreatedAt time.Time  `bson:"createdAt"`
	DeletedAt *time.Time `bson:"deletedAt,omitempty"`

	Params interface{} `bson:"params"`
}

type ResourceMigrationUpdateOwnerParams struct {
	OwnerID string `bson:"ownerId"`
}

// AsUpdateOwnerParams returns the params as an update owner params struct.
func (r *ResourceMigration) AsUpdateOwnerParams() *ResourceMigrationUpdateOwnerParams {
	if r.Type != ResourceMigrationTypeUpdateOwner {
		panic("cannot convert migration to update owner params")
	}

	return r.Params.(*ResourceMigrationUpdateOwnerParams)
}

// ToDTO returns the resource migration to a body.ResourceMigrationDTO
func (r *ResourceMigration) ToDTO() *body.ResourceMigrationRead {
	return &body.ResourceMigrationRead{
		ID:         r.ID,
		ResourceID: r.ResourceID,
		UserID:     r.UserID,

		Type:         r.Type,
		ResourceType: r.ResourceType,
		Status:       r.Status,

		CreatedAt: r.CreatedAt,
		DeletedAt: r.DeletedAt,
	}
}

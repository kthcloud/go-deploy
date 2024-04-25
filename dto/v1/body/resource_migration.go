package body

import "time"

type ResourceMigrationRead struct {
	ID         string `json:"id"`
	ResourceID string `json:"resourceId"`
	UserID     string `json:"userId"`

	Type         string `json:"type"`
	ResourceType string `json:"resourceType"`
	Status       string `json:"status"`

	CreatedAt time.Time  `json:"createdAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}

type ResourceMigrationCreate struct {
	Type string `json:"type" binding:"required"`

	ResourceID string `json:"resourceID" binding:"required,uuid4"`

	// TODO: this field is redundant, remove it
	ResourceType string `json:"resourceType" binding:"required"`

	UpdateOwner *struct {
		OwnerID string `json:"ownerId" binding:"required,uuid4"`
	} `json:"updateOwner,omitempty"`
}

type ResourceMigrationUpdate struct {
	Status string  `json:"status"`
	Token  *string `json:"token,omitempty"`
}

type ResourceMigrationCreated struct {
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}

type ResourceMigrationUpdated struct {
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}

type ResourceMigrationDeleted struct {
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}

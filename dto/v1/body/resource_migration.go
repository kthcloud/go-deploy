package body

import "time"

type ResourceMigrationRead struct {
	ID         string `json:"id"`
	ResourceID string `json:"resourceId"`
	UserID     string `json:"userId"`

	Type         string `json:"type"`
	ResourceType string `json:"resourceType"`
	Status       string `json:"status"`

	UpdateOwner *struct {
		OwnerID string `json:"ownerId"`
	} `json:"updateOwner,omitempty"`

	CreatedAt time.Time  `json:"createdAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}

type ResourceMigrationCreate struct {
	Type       string  `json:"type" binding:"required"`
	ResourceID string  `json:"resourceID" binding:"required,uuid4"`
	Status     *string `json:"status" binding:"required"`

	UpdateOwner *struct {
		OwnerID string `json:"ownerId" binding:"required,uuid4"`
	} `json:"updateOwner,omitempty"`
}

type ResourceMigrationUpdate struct {
	Status string  `json:"status" binding:"required"`
	Code   *string `json:"code,omitempty"`
}

type ResourceMigrationCreated struct {
	ResourceMigrationRead `json:",inline"`

	// JobID is the ID of the job that was created for the resource migration.
	// Only if the migration was created with status 'accepted' a job will be created.
	JobID *string `json:"jobId,omitempty"`
}

type ResourceMigrationUpdated struct {
	ResourceMigrationRead `json:",inline"`

	// JobID is the ID of the job that was created for the resource migration.
	// Only if the migration was updated with status 'accepted' a job will be created.
	JobID *string `json:"jobId,omitempty"`
}

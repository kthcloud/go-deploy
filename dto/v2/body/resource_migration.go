package body

import "time"

type ResourceMigrationRead struct {
	ID string `json:"id"`
	// ResourceID is the ID of the resource that is being migrated.
	// This can be a VM ID, deployment ID, etc. depending on the type of the migration.
	ResourceID string `json:"resourceId"`
	// UserID is the ID of the user who initiated the migration.
	UserID string `json:"userId"`

	// Type is the type of the resource migration.
	//
	// Possible values:
	// - updateOwner
	Type string `json:"type"`
	// ResourceType is the type of the resource that is being migrated.
	//
	// Possible values:
	// - vm
	// - deployment
	ResourceType string `json:"resourceType"`
	// Status is the status of the resource migration.
	// When this field is set to 'accepted', the migration will take place and then automatically be deleted.
	Status string `json:"status"`

	// UpdateOwner is the set of parameters that are required for the updateOwner migration type.
	// It is empty if the migration type is not updateOwner.
	UpdateOwner *struct {
		OwnerID string `json:"ownerId"`
	} `json:"updateOwner,omitempty"`

	CreatedAt time.Time  `json:"createdAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}

type ResourceMigrationCreate struct {
	// Type is the type of the resource migration.
	//
	// Possible values:
	// - updateOwner
	Type string `json:"type" binding:"required,oneof=updateOwner"`
	// ResourceID is the ID of the resource that is being migrated.
	// This can be a VM ID, deployment ID, etc. depending on the type of the migration.
	ResourceID string `json:"resourceId" binding:"required,uuid4"`
	// Status is the status of the resource migration.
	// It is used by privileged admins to directly accept or reject a migration.
	// The field is ignored by non-admins.
	//
	// Possible values:
	// - accepted
	// - pending
	Status *string `json:"status"`

	// UpdateOwner is the set of parameters that are required for the updateOwner migration type.
	// It is ignored if the migration type is not updateOwner.
	UpdateOwner *struct {
		OwnerID string `json:"ownerId" binding:"required,uuid4"`
	} `json:"updateOwner,omitempty"`
}

type ResourceMigrationUpdate struct {
	// Status is the status of the resource migration.
	// It is used to accept a migration by setting the status to 'accepted'.
	// If the acceptor is not an admin, a Code must be provided.
	//
	// Possible values:
	// - accepted
	// - pending
	Status string `json:"status" binding:"required"`
	// Code is a token required when accepting a migration if the acceptor is not an admin.
	// It is sent to the acceptor using the notification API
	Code *string `json:"code,omitempty"`
}

type ResourceMigrationCreated struct {
	ResourceMigrationRead `json:",inline" tstype:",extends"`

	// JobID is the ID of the job that was created for the resource migration.
	// It will only be set if the migration was created with status 'accepted'.
	JobID *string `json:"jobId,omitempty"`
}

type ResourceMigrationUpdated struct {
	ResourceMigrationRead `json:",inline" tstype:",extends"`

	// JobID is the ID of the job that was created for the resource migration.
	// It will only be set if the migration was updated with status 'accepted'.
	JobID *string `json:"jobId,omitempty"`
}

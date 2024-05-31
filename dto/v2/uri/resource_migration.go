package uri

type ResourceMigrationGet struct {
	ResourceMigrationID string `uri:"resourceMigrationId" binding:"required,uuid4"`
}

type ResourceMigrationUpdate struct {
	ResourceMigrationID string `uri:"resourceMigrationId" binding:"required,uuid4"`
}

type ResourceMigrationDelete struct {
	ResourceMigrationID string `uri:"resourceMigrationId" binding:"required,uuid4"`
}

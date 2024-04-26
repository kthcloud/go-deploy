package query

type VmGet struct {
	// MigrationCode is used when fetching a deployment that is being migrated.
	// The token should only be known by the user receiving the deployment.
	MigrationCode *string `json:"migrationCode" binding:"omitempty"`
}

type VmList struct {
	*Pagination

	All    bool    `form:"all" binding:"omitempty,boolean"`
	UserID *string `form:"userId" binding:"omitempty,uuid4"`
}

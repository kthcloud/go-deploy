package query

type VmGet struct {
	// MigrationToken is used when fetching a deployment that is being migrated.
	// The token should only be known by the user receiving the deployment.
	MigrationToken *string `json:"token" binding:"omitempty"`
}

type VmList struct {
	*Pagination

	All    bool    `form:"all" binding:"omitempty,boolean"`
	UserID *string `form:"userId" binding:"omitempty,uuid4"`
}

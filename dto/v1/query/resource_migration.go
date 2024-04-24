package query

type ResourceMigrationList struct {
	*Pagination
}

type ResourceMigrationUpdate struct {
	Token string `json:"token"`
}

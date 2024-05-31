package query

type JobList struct {
	*Pagination
	*SortBy

	All           bool     `form:"all" binding:"omitempty,boolean"`
	Status        []string `form:"status" binding:"omitempty"`
	ExcludeStatus []string `form:"excludeStatus" binding:"omitempty"`
	Types         []string `form:"type" binding:"omitempty"`
	ExcludeTypes  []string `form:"excludeType" binding:"omitempty"`
	UserID        *string  `form:"userId" binding:"omitempty"`
}

package query

type JobList struct {
	*Pagination
	*SortBy

	All    bool    `form:"all" binding:"omitempty,boolean"`
	Status *string `form:"status" binding:"omitempty,oneof=pending running failed terminated finished completed"`
	Type   *string `form:"type" binding:"omitempty,ascii"`
	UserID *string `form:"userId" binding:"omitempty,uuid4"`
}

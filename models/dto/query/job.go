package query

type JobGetMany struct {
	Pagination
	Status *string `form:"status" binding:"omitempty,oneof=pending running failed terminated finished"`
	Type   *string `form:"type" binding:"omitempty,ascii"`
	UserID *string `form:"userId" binding:"omitempty,uuid4"`
	All    bool    `form:"all" binding:"omitempty,boolean"`
}

package query

type JobGetMany struct {
	Pagination
	
	All    bool    `form:"all" binding:"omitempty,boolean"`
	Status *string `form:"status" binding:"omitempty,oneof=pending running failed terminated finished"`
	Type   *string `form:"type" binding:"omitempty,ascii"`
	UserID *string `form:"userId" binding:"omitempty,uuid4"`
}

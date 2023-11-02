package query

type VmList struct {
	Pagination

	All    bool    `form:"all" binding:"omitempty,boolean"`
	UserID *string `form:"userId" binding:"omitempty,uuid4"`
	Shared bool    `form:"shared" binding:"omitempty,boolean"`
}

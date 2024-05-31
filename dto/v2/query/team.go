package query

type TeamList struct {
	*Pagination

	UserID *string `form:"userId" binding:"omitempty"`
	All    bool    `form:"all" binding:"omitempty,boolean"`
}

package query

type UserList struct {
	Pagination

	UserID *string `form:"userId" binding:"omitempty,uuid"`
	All    bool    `form:"all" binding:"omitempty,boolean"`
}

type TeamList struct {
	Pagination

	UserID *string `form:"userId" binding:"omitempty,uuid"`
	All    bool    `form:"all" binding:"omitempty,boolean"`
}

package query

type UserList struct {
	Pagination

	All bool `form:"all" binding:"omitempty,boolean"`
}

type TeamList struct {
	Pagination

	UserID *string `form:"userId" binding:"omitempty,uuid"`
	All    bool    `form:"all" binding:"omitempty,boolean"`
}

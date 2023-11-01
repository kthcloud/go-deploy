package query

type UserList struct {
	Pagination

	UserID *string `form:"userId" binding:"omitempty,uuid"`
	All    bool    `form:"all" binding:"omitempty,boolean"`
	Search *string `form:"search" binding:"omitempty"`
}

type TeamList struct {
	Pagination

	UserID *string `form:"userId" binding:"omitempty,uuid"`
	All    bool    `form:"all" binding:"omitempty,boolean"`
}

package query

type UserList struct {
	*Pagination

	All      bool    `form:"all" binding:"omitempty,boolean"`
	Search   *string `form:"search" binding:"omitempty"`
	Discover bool    `form:"discover" binding:"omitempty,boolean"`
}

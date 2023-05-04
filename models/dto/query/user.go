package query

type UserList struct {
	WantAll bool `form:"all" binding:"omitempty,boolean"`
}

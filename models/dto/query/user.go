package query

type UserList struct {
	WantAll bool `query:"all" binding:"omitempty,boolean"`
}

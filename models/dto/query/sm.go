package query

type SmList struct {
	Pagination

	All bool `form:"all" binding:"omitempty,boolean"`
}

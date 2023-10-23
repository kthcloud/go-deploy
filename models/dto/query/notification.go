package query

type NotificationList struct {
	Pagination

	All    bool    `form:"all" binding:"omitempty,boolean"`
	UserID *string `form:"userId" binding:"omitempty,uuid4"`
}

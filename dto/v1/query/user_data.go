package query

type UserDataList struct {
	*Pagination

	UserID *string `form:"userId" binding:"omitempty,uuid4"`
}

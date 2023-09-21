package query

type Pagination struct {
	Page     int `form:"page" binding:"omitempty,min=0"`
	PageSize int `form:"pageSize" binding:"omitempty,min=0,max=10000"`
}

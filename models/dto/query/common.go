package query

type Pagination struct {
	Page     int `form:"page" binding:"omitempty,min=0"`
	PageSize int `form:"pageSize" binding:"omitempty,min=0,max=10000"`
}

type SortBy struct {
	Field string `form:"sortBy" binding:"omitempty"`
	Order int    `form:"sortOrder" binding:"omitempty,oneof=1 -1"`
}

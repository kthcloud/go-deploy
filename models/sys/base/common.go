package base

type Pagination struct {
	Page     int
	PageSize int
}

type SortBy struct {
	Field string
	Order int
}

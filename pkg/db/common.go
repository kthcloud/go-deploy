package db

type Pagination struct {
	Page     int
	PageSize int
}

type SortBy struct {
	Field string
	Order int
}

type SearchParams struct {
	Query  string   `json:"query"`
	Fields []string `json:"fields"`
}

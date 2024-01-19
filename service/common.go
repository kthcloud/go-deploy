package service

import "go-deploy/models/dto/query"

type Pagination struct {
	Page     int
	PageSize int
}

type SortBy struct {
	Field string
	Order int
}

func GetOrDefaultPagination(page *query.Pagination) *Pagination {
	if page == nil {
		return &Pagination{
			Page:     0,
			PageSize: 1000,
		}
	}

	if page.PageSize > 1000 {
		page.PageSize = 1000
	}

	return &Pagination{
		Page:     page.Page,
		PageSize: page.PageSize,
	}
}

func GetOrDefaultSortBy(sortBy *query.SortBy) *SortBy {
	if sortBy == nil {
		return &SortBy{
			Field: "createdAt",
			Order: 1,
		}
	}

	if sortBy.Field == "" {
		sortBy.Field = "createdAt"
	}

	if sortBy.Order == 0 {
		sortBy.Order = 1
	}

	return &SortBy{
		Field: sortBy.Field,
		Order: sortBy.Order,
	}
}

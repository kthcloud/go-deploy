package service

import "go-deploy/models/dto/query"

type Pagination struct {
	Page     int
	PageSize int
}

func GetOrDefault(page *query.Pagination) *Pagination {
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

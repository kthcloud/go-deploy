package query

type GpuClaimList struct {
	*Pagination
	Detailed bool `form:"detailed"`
}

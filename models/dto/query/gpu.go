package query

type GpuList struct {
	OnlyShowAvailable bool `form:"available" binding:"omitempty,boolean"`

	Pagination
}

package query

type GpuList struct {
	OnlyShowAvailable bool    `form:"available" binding:"omitempty,boolean"`
	Zone              *string `form:"zone" binding:"omitempty"`

	Pagination
}

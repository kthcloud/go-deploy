package query

type VmList struct {
	WantAll bool `form:"all" binding:"omitempty,boolean"`
}

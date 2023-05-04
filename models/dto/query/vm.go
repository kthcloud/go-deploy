package query

type VmList struct {
	WantAll bool `query:"all" binding:"omitempty,boolean"`
}

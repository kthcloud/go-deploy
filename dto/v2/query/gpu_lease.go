package query

type GpuLeaseList struct {
	*Pagination

	All   bool     `form:"all" binding:"omitempty,boolean"`
	VmIDs []string `form:"vmId" binding:"omitempty"`
}

type GpuLeaseCreate struct {
	VmID string `form:"vmId" binding:"required,uuid4"`
}

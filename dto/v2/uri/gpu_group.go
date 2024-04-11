package uri

type GpuGroupList struct {
}

type GpuGroupGet struct {
	GpuGroupID string `uri:"gpuGroupId" binding:"required"`
}

type GpuGroupDelete struct {
	GpuGroupID string `uri:"gpuGroupId" binding:"required"`
}

type GpuGroupUpdate struct {
	GpuGroupID string `uri:"gpuGroupId" binding:"required"`
}

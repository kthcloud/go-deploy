package uri

type GpuClaimList struct {
}

type GpuClaimGet struct {
	GpuClaimID string `uri:"gpuClaimId" binding:"required"`
}

type GpuClaimDelete struct {
	GpuClaimID string `uri:"gpuClaimId" binding:"required"`
}

package uri

type GpuLeaseList struct {
}

type GpuLeaseGet struct {
	GpuLeaseID string `uri:"gpuLeaseId" binding:"required,uuid4"`
}

type GpuLeaseDelete struct {
	GpuLeaseID string `uri:"gpuLeaseId" binding:"required,uuid4"`
}

type GpuLeaseUpdate struct {
	GpuLeaseID string `uri:"gpuLeaseId" binding:"required,uuid4"`
}

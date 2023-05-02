package uri

type GpuGet struct {
	GpuID string `uri:"gpuId" binding:"required,uuid4"`
}

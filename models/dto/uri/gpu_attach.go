package uri

type GpuAttach struct {
	VmID  string `uri:"vmId" binding:"required,uuid4"`
	GpuID string `uri:"gpuId" binding:"omitempty,base64"`
}

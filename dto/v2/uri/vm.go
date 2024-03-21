package uri

type VmGet struct {
	VmID string `uri:"vmId" binding:"required,uuid4"`
}

type VmDelete struct {
	VmID string `uri:"vmId" binding:"required,uuid4"`
}

type VmUpdate struct {
	VmID string `uri:"vmId" binding:"required,uuid4"`
}

type GpuAttach struct {
	VmID  string `uri:"vmId" binding:"required,uuid4"`
	GpuID string `uri:"gpuId" binding:"omitempty,base64"`
}

type GpuDetach struct {
	VmID string `uri:"vmId" binding:"required,uuid4"`
}

type GpuGet struct {
	GpuID string `uri:"gpuId" binding:"required,uuid4"`
}

type VmCommand struct {
	VmID string `uri:"vmId" binding:"required,uuid4"`
}

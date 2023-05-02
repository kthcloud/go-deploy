package uri

type GpuDetach struct {
	VmID string `uri:"vmId" binding:"required,uuid4"`
}

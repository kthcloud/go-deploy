package uri

type VmGet struct {
	VmID string `uri:"vmId" binding:"required,uuid4"`
}

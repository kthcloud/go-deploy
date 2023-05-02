package uri

type VmDelete struct {
	VmID string `uri:"vmId" binding:"required,uuid4"`
}

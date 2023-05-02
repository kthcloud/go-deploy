package uri

type DoCommand struct {
	VmID string `uri:"vmId" binding:"required,uuid4"`
}

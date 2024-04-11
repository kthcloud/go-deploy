package query

type VmActionCreate struct {
	VmID string `form:"vmId" binding:"required,uuid4"`
}

package body

type VmActionCreate struct {
	Action string `json:"action" binding:"required,oneof=start stop restart repair"`
}

type VmActionCreated struct {
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}

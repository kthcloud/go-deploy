package body

type VmAction struct {
	Action string `json:"action" binding:"required,oneof=start stop restart repair"`
}

type VmActionDone struct {
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}

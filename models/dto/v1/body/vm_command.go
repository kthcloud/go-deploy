package body

type VmCommand struct {
	Command string `json:"command" binding:"required,oneof=start stop reboot"`
}

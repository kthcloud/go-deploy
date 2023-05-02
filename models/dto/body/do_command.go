package body

type DoCommand struct {
	Command string `json:"command" binding:"required,oneof=start stop reboot"`
}

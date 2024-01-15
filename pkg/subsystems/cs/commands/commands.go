package commands

type Command string

const (
	// Start used to start a VM
	Start Command = "start"
	// Stop used to stop a VM
	Stop Command = "stop"
	// Reboot used to reboot a VM
	Reboot Command = "reboot"
	// Restart used to restart a VM
	Restart Command = "restart"
)

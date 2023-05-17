package commands

type Command string

const (
	Start   Command = "start"
	Stop    Command = "stop"
	Reboot  Command = "reboot"
	Restart Command = "restart"
)

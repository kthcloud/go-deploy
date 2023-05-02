package body

type VmCreate struct {
	Name         string `json:"name" binding:"required,rfc1035,min=3,max=30"`
	SshPublicKey string `json:"sshPublicKey" binding:"required,ssh_public_key"`
}

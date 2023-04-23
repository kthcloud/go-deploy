package dto

type VmCreate struct {
	Name         string `json:"name"`
	SshPublicKey string `json:"sshPublicKey"`
}

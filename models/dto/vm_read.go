package dto

type VmRead struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	SshPublicKey     string `json:"sshPublicKey"`
	OwnerID          string `json:"ownerId"`
	Status           string `json:"status"`
	ConnectionString string `json:"connectionString"`
}

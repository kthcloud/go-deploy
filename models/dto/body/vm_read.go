package body

import "time"

type VmGpu struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	LeaseEnd time.Time `json:"leaseEnd"`
}

type VmRead struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	SshPublicKey     string `json:"sshPublicKey"`
	OwnerID          string `json:"ownerId"`
	Status           string `json:"status"`
	ConnectionString string `json:"connectionString"`
	GPU              *VmGpu `json:"gpu,omitempty"`
}

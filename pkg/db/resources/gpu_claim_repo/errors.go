package gpu_claim_repo

import "errors"

var (
	// ErrGpuClaimAlreadyExists is returned when a GPU Claim already exists with the same name
	ErrGpuClaimAlreadyExists = errors.New("gpu claim already exists")
)

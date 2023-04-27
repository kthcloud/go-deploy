package vm_service

import (
	vmModel "go-deploy/models/vm"
	"time"
)

// GetAllGPUs TODO: add filter for isGpuUser
func GetAllGPUs(showOnlyAvailable bool, isGpuUser bool) ([]vmModel.GPU, error) {
	if showOnlyAvailable {
		return vmModel.GetAllAvailableGPUs()
	}
	return vmModel.GetAllGPUs()
}

// GetGpuByID TODO: add filter for isGpuUser
func GetGpuByID(gpuID string, isGpuUser bool) (*vmModel.GPU, error) {
	return vmModel.GetGpuByID(gpuID)
}

func AttachGpuToVM(gpuID, vmID, userID string) (bool, error) {
	// TODO: add check for user's quota
	oneHourFromNow := time.Now().Add(time.Hour)

	return vmModel.AttachGpuToVM(gpuID, vmID, userID, oneHourFromNow)
}

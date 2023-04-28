package vm_service

import (
	vmModel "go-deploy/models/vm"
	"go-deploy/service/vm_service/internal_service"
	"log"
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
	if gpuID == "any" {
		return vmModel.GetAnyAvailableGPU()
	}

	return vmModel.GetGpuByID(gpuID)
}

func AttachGPU(gpuID, vmID, userID string) {
	go func() {
		// TODO: add check for user's quota
		oneHourFromNow := time.Now().Add(time.Hour)

		attached, err := vmModel.AttachGPU(gpuID, vmID, userID, oneHourFromNow)
		if err != nil {
			log.Println(err)
			return
		}

		if !attached {
			log.Println("did not attach gpu", gpuID, "to vm", vmID)
			return
		}

		err = internal_service.AttachGPU(gpuID, vmID)
		if err != nil {
			log.Println(err)
			return
		}

	}()
}

func DetachGPU(vmID, userID string) {
	go func() {
		err := internal_service.DetachGPU(vmID)
		if err != nil {
			log.Println(err)
			return
		}

		detached, err := vmModel.DetachGPU(vmID, userID)
		if err != nil {
			log.Println(err)
			return
		}

		if !detached {
			log.Println("did not detach gpu from vm", vmID)
			return
		}
	}()
}

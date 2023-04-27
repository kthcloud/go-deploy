package vm_service

import vmModel "go-deploy/models/vm"

func GetAllGPUs(showOnlyAvailable bool) ([]vmModel.GPU, error) {
	if showOnlyAvailable {
		return vmModel.GetAllAvailableGPUs()
	}
	return vmModel.GetAllGPUs()
}

func GetAllBasicGPUs(showOnlyAvailable bool) ([]vmModel.GPU, error) {
	// TODO: apply filter to get only basic GPUs
	if showOnlyAvailable {
		return vmModel.GetAllAvailableGPUs()
	}
	return vmModel.GetAllGPUs()
}

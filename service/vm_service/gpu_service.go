package vm_service

import (
	"fmt"
	vmModel "go-deploy/models/vm"
	"go-deploy/pkg/conf"
	"go-deploy/service/vm_service/internal_service"
	"log"
	"sort"
	"time"
)

func GetAllGPUs(showOnlyAvailable bool, isPowerUser bool) ([]vmModel.GPU, error) {
	var excludedGPUs []string
	if !isPowerUser {
		excludedGPUs = conf.Env.GPU.PrivilegedGPUs
	}

	if showOnlyAvailable {
		return vmModel.GetAllAvailableGPUs(conf.Env.GPU.ExcludedHosts, excludedGPUs)
	}
	return vmModel.GetAllGPUs()
}

func GetGpuByID(gpuID string, isPowerUser bool) (*vmModel.GPU, error) {
	gpu, err := vmModel.GetGpuByID(gpuID)
	if err != nil {
		return nil, err
	}

	if gpu == nil {
		return nil, nil
	}

	// check if host is excluded
	for _, excludedHost := range conf.Env.GPU.ExcludedHosts {
		if gpu.Host == excludedHost {
			return nil, nil
		}
	}

	if !isPowerUser {
		// check if card is privileged
		if isGpuPrivileged(gpu.Data.Name) {
			return nil, nil
		}
	}

	return gpu, nil
}

func IsGpuAvailable(gpu *vmModel.GPU) (bool, error) {
	// check if attached in cloudstack
	attached, err := internal_service.IsGpuAttachedCS(gpu.Host, gpu.Data.Bus)
	if err != nil {
		return false, err
	}

	if attached {
		return false, nil
	}

	return true, nil
}

func GetAnyAvailableGPU(isPowerUser bool) (*vmModel.GPU, error) {
	var excludedGPUs []string
	if !isPowerUser {
		excludedGPUs = conf.Env.GPU.PrivilegedGPUs
	}

	availableGPUs, err := vmModel.GetAllAvailableGPUs(conf.Env.GPU.ExcludedHosts, excludedGPUs)
	if err != nil {
		return nil, err
	}

	// sort available gpus by host
	sort.Slice(availableGPUs, func(i, j int) bool {
		return availableGPUs[i].Host < availableGPUs[j].Host
	})

	for _, gpu := range availableGPUs {
		// check if attached in cloudstack
		inUse, err := isGpuInUse(gpu.Host, gpu.Data.Bus)
		if err != nil {
			return nil, err
		}

		if !inUse {
			return &gpu, nil
		}
	}

	return nil, nil
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
		err := DetachGpuSync(vmID, userID)
		if err != nil {
			log.Println(err)
		}
	}()
}

func DetachGpuSync(vmID, userID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to detach gpu from vm %s. details: %s", vmID, err)
	}

	err := internal_service.DetachGPU(vmID)
	if err != nil {
		return makeError(err)
	}

	detached, err := vmModel.DetachGPU(vmID, userID)
	if err != nil {
		return makeError(err)
	}

	if !detached {
		return makeError(fmt.Errorf("failed to detach gpu from vm %s", vmID))
	}

	return nil
}

func isGpuPrivileged(cardName string) bool {
	for _, privilegedCard := range conf.Env.GPU.PrivilegedGPUs {
		if cardName == privilegedCard {
			return true
		}
	}

	return false
}

func isHostExcluded(hostName string) bool {
	for _, excludedHost := range conf.Env.GPU.ExcludedHosts {
		if hostName == excludedHost {
			return true
		}
	}

	return false
}

func isGpuInUse(hostName, bus string) (bool, error) {
	attached, err := internal_service.IsGpuAttachedCS(hostName, bus)
	if err != nil {
		return false, err
	}

	if attached {
		return true, nil
	}

	return false, nil
}

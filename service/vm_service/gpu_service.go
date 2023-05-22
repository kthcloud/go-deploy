package vm_service

import (
	"errors"
	"fmt"
	vmModel "go-deploy/models/sys/vm"
	gpuModel "go-deploy/models/sys/vm/gpu"
	"go-deploy/pkg/conf"
	"go-deploy/service/vm_service/internal_service"
	"sort"
	"strings"
	"time"
)

func GetAllGPUs(showOnlyAvailable bool, isPowerUser bool) ([]gpuModel.GPU, error) {
	var excludedGPUs []string
	if !isPowerUser {
		excludedGPUs = conf.Env.GPU.PrivilegedGPUs
	}

	if showOnlyAvailable {
		return gpuModel.GetAllAvailableGPUs(conf.Env.GPU.ExcludedHosts, excludedGPUs)
	}
	return gpuModel.GetAllGPUs(conf.Env.GPU.ExcludedHosts, excludedGPUs)
}

func GetGpuByID(gpuID string, isPowerUser bool) (*gpuModel.GPU, error) {
	gpu, err := gpuModel.GetGpuByID(gpuID)
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

func IsGpuAvailable(gpu *gpuModel.GPU) (bool, error) {
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

func GetAnyAvailableGPU(isPowerUser bool) (*gpuModel.GPU, error) {
	var excludedGPUs []string
	if !isPowerUser {
		excludedGPUs = conf.Env.GPU.PrivilegedGPUs
	}

	availableGPUs, err := gpuModel.GetAllAvailableGPUs(conf.Env.GPU.ExcludedHosts, excludedGPUs)
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

func GetAllAvailableGPU(isPowerUser bool) ([]gpuModel.GPU, error) {
	var excludedGPUs []string
	if !isPowerUser {
		excludedGPUs = conf.Env.GPU.PrivilegedGPUs
	}

	availableGPUs, err := gpuModel.GetAllAvailableGPUs(conf.Env.GPU.ExcludedHosts, excludedGPUs)
	if err != nil {
		return nil, err
	}

	// sort available gpus by host
	sort.Slice(availableGPUs, func(i, j int) bool {
		return availableGPUs[i].Host < availableGPUs[j].Host
	})

	var notInUseGPUs []gpuModel.GPU
	for _, gpu := range availableGPUs {
		// check if attached in cloudstack
		inUse, err := isGpuInUse(gpu.Host, gpu.Data.Bus)
		if err != nil {
			return nil, err
		}

		if !inUse {
			notInUseGPUs = append(notInUseGPUs, gpu)
		}
	}

	return notInUseGPUs, nil
}

func AttachGPU(gpuIDs []string, vmID, userID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to attach gpu to vm %s. details: %s", vmID, err)
	}
	csInsufficientCapacityError := "host has capacity? false"

	// TODO: add check for user's quota
	oneHourFromNow := time.Now().Add(time.Hour)

	var err error
	for idx, gpuID := range gpuIDs {
		var attached bool
		attached, err = gpuModel.AttachGPU(gpuID, vmID, userID, oneHourFromNow)
		if err != nil {
			return makeError(err)
		}

		if !attached {
			// this is an edge case where we don't want to fail the method, since a retry will probably not help
			//
			// this is probably caused by a race condition where two users requested the same gpu, where the first one
			// got it, and the second one failed. we don't want to fail the second user, since that would mean that a
			// job would get stuck. instead the user is not granted the gpu, and will need to request a new one manually
			continue
		}

		err = internal_service.AttachGPU(gpuID, vmID)
		if err == nil {
			break
		}

		errString := err.Error()
		if strings.Contains(errString, csInsufficientCapacityError) {
			// if the host has insufficient capacity, we need to detach the gpu from the vm
			// and attempt to attach it to another gpu, if we have any left to try

			if idx == len(gpuIDs)-1 {
				err = makeError(errors.New("insufficient capacity on host"))
				break
			}

			err = gpuModel.DetachGPU(vmID, userID)
			if err != nil {
				return makeError(err)
			}

			err = internal_service.DetachGPU(vmID)
			if err != nil {
				return makeError(err)
			}
		} else {
			return makeError(err)
		}
	}

	err = vmModel.RemoveActivity(vmID, vmModel.ActivityAttachingGPU)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func DetachGPU(vmID, userID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to detach gpu from vm %s. details: %s", vmID, err)
	}

	err := DetachGpuSync(vmID, userID)
	if err != nil {
		return makeError(err)
	}

	err = vmModel.RemoveActivity(vmID, vmModel.ActivityDetachingGPU)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func DetachGpuSync(vmID, userID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to detach gpu from vm %s. details: %s", vmID, err)
	}

	err := internal_service.DetachGPU(vmID)
	if err != nil {
		return makeError(err)
	}

	err = gpuModel.DetachGPU(vmID, userID)
	if err != nil {
		return makeError(err)
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

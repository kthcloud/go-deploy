package vm_service

import (
	"errors"
	"fmt"
	gpuModel "go-deploy/models/sys/gpu"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/config"
	"go-deploy/service"
	"go-deploy/service/vm_service/cs_service"
	"go-deploy/service/vm_service/service_errors"
	"go-deploy/utils"
	"log"
	"sort"
	"strings"
	"time"
)

func ListGPUs(onlyAvailable bool, auth *service.AuthInfo) ([]gpuModel.GPU, error) {
	effectiveRole := auth.GetEffectiveRole()

	excludedGPUs := config.Config.GPU.ExcludedGPUs
	if !effectiveRole.Permissions.UsePrivilegedGPUs {
		excludedGPUs = append(excludedGPUs, config.Config.GPU.PrivilegedGPUs...)
	}

	if onlyAvailable {
		dbAvailableGPUs, err := gpuModel.NewWithExclusion(config.Config.GPU.ExcludedHosts, excludedGPUs).GetAllAvailable()
		if err != nil {
			return nil, err
		}

		availableGPUs := make([]gpuModel.GPU, 0)
		for _, gpu := range dbAvailableGPUs {
			err = IsGpuHardwareAvailable(&gpu)
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error checking if gpu is in use. details: %w", err))
				continue
			}

			availableGPUs = append(availableGPUs, gpu)
		}

		return availableGPUs, nil
	}
	return gpuModel.NewWithExclusion(config.Config.GPU.ExcludedHosts, excludedGPUs).List()
}

func GetGpuByID(gpuID string, usePrivilegedGPUs bool) (*gpuModel.GPU, error) {
	gpu, err := gpuModel.New().GetByID(gpuID)
	if err != nil {
		return nil, err
	}

	if gpu == nil {
		return nil, nil
	}

	// check if host is excluded
	for _, excludedHost := range config.Config.GPU.ExcludedHosts {
		if gpu.Host == excludedHost {
			return nil, nil
		}
	}

	if !usePrivilegedGPUs {
		// check if card is privileged
		if isGpuPrivileged(gpu.Data.Name) {
			return nil, nil
		}
	}

	return gpu, nil
}

func GetAllAvailableGPU(usePrivilegedGPUs bool) ([]gpuModel.GPU, error) {
	excludedGPUs := config.Config.GPU.ExcludedGPUs
	if !usePrivilegedGPUs {
		excludedGPUs = append(excludedGPUs, config.Config.GPU.PrivilegedGPUs...)
	}

	dbAvailableGPUs, err := gpuModel.NewWithExclusion(config.Config.GPU.ExcludedHosts, excludedGPUs).GetAllAvailable()
	if err != nil {
		return nil, err
	}

	// sort available gpus by host
	sort.Slice(dbAvailableGPUs, func(i, j int) bool {
		return dbAvailableGPUs[i].Host < dbAvailableGPUs[j].Host
	})

	var availableGPUs []gpuModel.GPU
	for _, gpu := range dbAvailableGPUs {
		err = IsGpuHardwareAvailable(&gpu)
		if err != nil {
			return nil, err
		}

		availableGPUs = append(availableGPUs, gpu)
	}

	return availableGPUs, nil
}

func IsGpuPrivileged(gpuID string) (bool, error) {
	gpu, err := gpuModel.New().GetByID(gpuID)
	if err != nil {
		return false, err
	}

	if gpu == nil {
		return false, nil
	}

	for _, privilegedGPU := range config.Config.GPU.PrivilegedGPUs {
		if privilegedGPU == gpu.Data.Name {
			return true, nil
		}
	}

	return false, nil
}

func AttachGPU(gpuIDs []string, vmID, userID string, leaseDuration float64) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to attach gpu to vm %s. details: %w", vmID, err)
	}
	csInsufficientCapacityError := "host has capacity? false"
	gpuAlreadyAttachedError := "Unable to create a deployment for VM instance"

	endLease := time.Now().Add(time.Duration(leaseDuration) * time.Hour)

	var err error
	for _, gpuID := range gpuIDs {
		var gpu *gpuModel.GPU
		gpu, err = gpuModel.New().GetByID(gpuID)
		if err != nil {
			return makeError(err)
		}

		if gpu == nil {
			continue
		}

		requiresCsAttach := gpu.Lease.VmID != vmID || gpu.Lease.VmID == ""

		err = IsGpuHardwareAvailable(gpu)
		if err != nil {
			return makeError(err)
		}

		if gpu.Lease.VmID != vmID && gpu.IsAttached() {
			// if it is attached but expired, take over the card by first detaching it
			if gpu.Lease.IsExpired() {
				err = cs_service.DetachGPU(gpu.Lease.VmID, cs_service.CsDetachGpuAfterStateRestore)
				if err != nil {
					return makeError(err)
				}

				err = gpuModel.New().Detach(gpu.Lease.VmID)
				if err != nil {
					return makeError(err)
				}
			} else {
				continue
			}
		}

		var attached bool
		attached, err = gpuModel.New().Attach(gpuID, vmID, userID, endLease)
		if err != nil {
			if errors.Is(err, gpuModel.AlreadyAttachedErr) || errors.Is(err, gpuModel.NotFoundErr) {
				// this is not treated as an error, just another instance snatched the gpu before this one
				continue
			}

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

		if requiresCsAttach {
			err = cs_service.AttachGPU(gpuID, vmID)
			if err == nil {
				break
			}

			errString := err.Error()

			insufficientCapacityErr := strings.Contains(errString, csInsufficientCapacityError)
			gpuAlreadyAttached := strings.Contains(errString, gpuAlreadyAttachedError)

			if insufficientCapacityErr {
				// if the host has insufficient capacity, we need to detach the gpu from the vm
				// and attempt to attach it to another gpu

				err = cs_service.DetachGPU(vmID, cs_service.CsDetachGpuAfterStateRestore)
				if err != nil {
					return makeError(err)
				}

				err = gpuModel.New().Detach(vmID)
				if err != nil {
					return makeError(err)
				}
			} else if gpuAlreadyAttached {
				// if the gpu is already attached, we need to detach it from the vm

				err = cs_service.DetachGPU(vmID, cs_service.CsDetachGpuAfterStateRestore)
				if err != nil {
					return makeError(err)
				}

				err = gpuModel.New().Detach(vmID)
				if err != nil {
					return makeError(err)
				}
			} else {
				return makeError(err)
			}
		}
	}

	return nil
}

func DetachGPU(vmID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to detach gpu from vm %s. details: %w", vmID, err)
	}

	err := DetachGpuSync(vmID)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func DetachGpuSync(vmID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to detach gpu from vm %s. details: %w", vmID, err)
	}

	err := cs_service.DetachGPU(vmID, cs_service.CsDetachGpuAfterStateRestore)
	if err != nil {
		return makeError(err)
	}

	err = gpuModel.New().Detach(vmID)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func CanStartOnHost(vmID, host string) error {
	vm, err := vmModel.New().GetByID(vmID)
	if err != nil {
		return err
	}

	if vm == nil {
		return service_errors.VmNotFoundErr
	}

	if vm.Subsystems.CS.VM.ID == "" {
		return service_errors.VmNotCreatedErr
	}

	err = cs_service.CanStart(vm.Subsystems.CS.VM.ID, host, vm.Zone)
	if err != nil {
		return err
	}

	return nil
}

func IsGpuHardwareAvailable(gpu *gpuModel.GPU) error {
	cloudstackAttached, err := cs_service.IsGpuAttachedCS(gpu)
	if err != nil {
		return err
	}

	zone := config.Config.VM.GetZone(gpu.Zone)
	if zone == nil {
		return service_errors.ZoneNotFoundErr
	}

	err = cs_service.HostInCorrectState(gpu.Host, zone)
	if err != nil {
		return err
	}

	// check if it is a "bad attach", where cloudstack reports it being attached, but the database says it's not
	// this usually means it is in use outside the scope of deploy
	if cloudstackAttached && !gpu.IsAttached() {
		return service_errors.GpuNotFoundErr
	}

	return nil
}

func isGpuPrivileged(cardName string) bool {
	for _, privilegedCard := range config.Config.GPU.PrivilegedGPUs {
		if cardName == privilegedCard {
			return true
		}
	}

	return false
}

func hasExtraConfig(vm *vmModel.VM) bool {
	if vm.Subsystems.CS.VM.ID == "" {
		log.Println("cs vm not found when checking for extra config when repairing gpus for vm", vm.ID, ". assuming it was deleted")
		return false
	}

	return vm.Subsystems.CS.VM.ExtraConfig != "" && vm.Subsystems.CS.VM.ExtraConfig != "none"
}

package vm_service

import (
	"fmt"
	vmModel "go-deploy/models/sys/vm"
	gpuModel "go-deploy/models/sys/vm/gpu"
	"go-deploy/pkg/conf"
	"go-deploy/service/vm_service/internal_service"
	"go-deploy/utils"
	"log"
	"sort"
	"strings"
	"time"
)

func GetAllGPUs(onlyAvailable bool, usePrivilegedGPUs bool) ([]gpuModel.GPU, error) {
	excludedGPUs := conf.Env.GPU.ExcludedGPUs
	if !usePrivilegedGPUs {
		excludedGPUs = append(excludedGPUs, conf.Env.GPU.PrivilegedGPUs...)
	}

	if onlyAvailable {
		dbAvailableGPUs, err := gpuModel.GetAllAvailable(conf.Env.GPU.ExcludedHosts, excludedGPUs)
		if err != nil {
			return nil, err
		}

		availableGPUs := make([]gpuModel.GPU, 0)
		for _, gpu := range dbAvailableGPUs {
			hardwareAvailable, err := IsGpuHardwareAvailable(&gpu)
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error checking if gpu is in use. details: %w", err))
				continue
			}

			if hardwareAvailable {
				availableGPUs = append(availableGPUs, gpu)
			}
		}

		return availableGPUs, nil
	}
	return gpuModel.GetAll(conf.Env.GPU.ExcludedHosts, excludedGPUs)
}

func GetGpuByID(gpuID string, usePrivilegedGPUs bool) (*gpuModel.GPU, error) {
	gpu, err := gpuModel.GetByID(gpuID)
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

	if !usePrivilegedGPUs {
		// check if card is privileged
		if isGpuPrivileged(gpu.Data.Name) {
			return nil, nil
		}
	}

	return gpu, nil
}

func GetAllAvailableGPU(usePrivilegedGPUs bool) ([]gpuModel.GPU, error) {
	excludedGPUs := conf.Env.GPU.ExcludedGPUs
	if !usePrivilegedGPUs {
		excludedGPUs = append(excludedGPUs, conf.Env.GPU.PrivilegedGPUs...)
	}

	dbAvailableGPUs, err := gpuModel.GetAllAvailable(conf.Env.GPU.ExcludedHosts, excludedGPUs)
	if err != nil {
		return nil, err
	}

	// sort available gpus by host
	sort.Slice(dbAvailableGPUs, func(i, j int) bool {
		return dbAvailableGPUs[i].Host < dbAvailableGPUs[j].Host
	})

	var availableGPUs []gpuModel.GPU
	for _, gpu := range dbAvailableGPUs {
		hardwareAvailable, err := IsGpuHardwareAvailable(&gpu)
		if err != nil {
			return nil, err
		}

		if hardwareAvailable {
			availableGPUs = append(availableGPUs, gpu)
		}
	}

	return availableGPUs, nil
}

func IsGpuPrivileged(gpuID string) (bool, error) {
	gpu, err := gpuModel.GetByID(gpuID)
	if err != nil {
		return false, err
	}

	if gpu == nil {
		return false, nil
	}

	for _, privilegedGPU := range conf.Env.GPU.PrivilegedGPUs {
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

	endLease := time.Now().Add(time.Duration(leaseDuration) * time.Hour)

	var err error
	for _, gpuID := range gpuIDs {
		var gpu *gpuModel.GPU
		gpu, err = gpuModel.GetByID(gpuID)
		if err != nil {
			return makeError(err)
		}

		if gpu == nil {
			continue
		}

		var hardwareAvailable bool
		hardwareAvailable, err = IsGpuHardwareAvailable(gpu)
		if err != nil {
			return makeError(err)
		}

		if !hardwareAvailable {
			continue
		}

		if gpu.Lease.VmID != vmID && gpu.IsAttached() {
			// if it is attached but expired, take over the card by first detaching it
			if gpu.Lease.IsExpired() {
				err = internal_service.DetachGPU(gpu.Lease.VmID, internal_service.CsDetachGpuAfterStateRestore)
				if err != nil {
					return makeError(err)
				}

				err = gpuModel.Detach(gpu.Lease.VmID, gpu.Lease.UserID)
				if err != nil {
					return makeError(err)
				}
			} else {
				continue
			}
		}

		var attached bool
		attached, err = gpuModel.Attach(gpuID, vmID, userID, endLease)
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
			// and attempt to attach it to another gpu

			err = internal_service.DetachGPU(vmID, internal_service.CsDetachGpuAfterStateRestore)
			if err != nil {
				return makeError(err)
			}

			err = gpuModel.Detach(vmID, userID)
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

func RepairGPUs() error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair gpus. details: %w", err)
	}

	// get all gpus that are attached to a vm
	attachedGPUs, err := gpuModel.GetAllLeased(nil, nil)
	if err != nil {
		return makeError(err)
	}

	// get all vms with an assigned gpu
	vmsWithGPU, err := vmModel.GetWithGPU()
	if err != nil {
		return makeError(err)
	}

	// find vm with same gpu
	gpuToVM := make(map[string]*vmModel.VM)
	for _, vm := range vmsWithGPU {
		_, ok := gpuToVM[vm.GpuID]
		if ok {
			vm1 := gpuToVM[vm.GpuID]
			vm2 := vm
			log.Println("found two vms with the same gpu. vm1:", vm1.ID, "(", vm1.Name, ")  vm2:", vm2.ID, "(", vm2.Name, "), gpu: ", vm.GpuID, ". detaching gpu from vm2")
			err = DetachGPU(vm2.ID, vm2.OwnerID)
			if err != nil {
				return makeError(err)
			}
		}

		gpuToVM[vm.GpuID] = &vm
	}

	// find gpus that are attached to a vm, but not in the gpuToVM map
	for _, gpu := range attachedGPUs {
		_, ok := gpuToVM[gpu.ID]
		if !ok {
			log.Println("found gpu that is attached to a vm, but not in the gpuToVM map. clearing lease. vm:", gpu.Lease.VmID, "gpu:", gpu.ID, "("+gpu.Data.Name+")")
			err = gpuModel.ClearLease(gpu.ID)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// find vms that have a gpu assigned, but the gpu has no lease
	for _, vm := range vmsWithGPU {
		_, ok := gpuToVM[vm.GpuID]
		if ok {
			matched := false
			for _, gpu := range attachedGPUs {
				if gpu.ID == vm.GpuID {
					matched = true
					break
				}
			}
			if matched {
				continue
			}

			// detach gpu from vm since we don't know how long it should be leased for
			log.Println("found vm that has a gpu assigned, but the gpu has no lease. detaching gpu from vm:", vm.ID, "(", vm.Name, ")")
			err = DetachGPU(vm.ID, vm.OwnerID)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// find vms that have a gpu assigned, but cs is not setup to use it
	for _, vm := range vmsWithGPU {
		if !hasExtraConfig(&vm) {
			err := internal_service.AttachGPU(vm.GpuID, vm.ID)
			if err != nil {
				return makeError(err)
			}
		}
	}

	log.Println("successfully repaired gpus")
	return nil
}

func DetachGPU(vmID, userID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to detach gpu from vm %s. details: %w", vmID, err)
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
		return fmt.Errorf("failed to detach gpu from vm %s. details: %w", vmID, err)
	}

	err := internal_service.DetachGPU(vmID, internal_service.CsDetachGpuAfterStateRestore)
	if err != nil {
		return makeError(err)
	}

	err = gpuModel.Detach(vmID, userID)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func CanStartOnHost(vmID, host string) (bool, string, error) {
	vm, err := vmModel.GetByID(vmID)
	if err != nil {
		return false, "", err
	}

	if vm == nil {
		return false, "", fmt.Errorf("vm %s not found", vmID)
	}

	if vm.Subsystems.CS.VM.ID == "" {
		return false, "VM not fully created", nil
	}

	canStart, reason, err := internal_service.CanStartCS(vm.Subsystems.CS.VM.ID, host, vm.Zone)
	if err != nil {
		return false, "", err
	}

	if !canStart {
		return false, reason, nil
	}

	return true, "", nil
}

func IsGpuHardwareAvailable(gpu *gpuModel.GPU) (bool, error) {
	cloudstackAttached, err := internal_service.IsGpuAttachedCS(gpu)
	if err != nil {
		return false, err
	}

	zone := conf.Env.VM.GetZone(gpu.Zone)
	if zone == nil {
		return false, fmt.Errorf("zone %s not found", gpu.Zone)
	}

	correctState, _, err := internal_service.HostInCorrectState(gpu.Host, zone)
	if err != nil {
		return false, err
	}

	// a "bad attach" is when cloudstack reports it being bound, but the database says it's not
	// this usually means it is in use outside the scope of deploy
	badAttach := cloudstackAttached && !gpu.IsAttached()

	return !badAttach && correctState, nil
}

func isGpuPrivileged(cardName string) bool {
	for _, privilegedCard := range conf.Env.GPU.PrivilegedGPUs {
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

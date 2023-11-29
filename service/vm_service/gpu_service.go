package vm_service

import (
	"fmt"
	vmModel "go-deploy/models/sys/vm"
	gpuModel "go-deploy/models/sys/vm/gpu"
	"go-deploy/pkg/config"
	"go-deploy/service/vm_service/cs_service"
	"go-deploy/utils"
	"log"
	"sort"
	"strings"
	"time"
)

func ListGPUs(onlyAvailable bool, usePrivilegedGPUs bool) ([]gpuModel.GPU, error) {
	excludedGPUs := config.Config.GPU.ExcludedGPUs
	if !usePrivilegedGPUs {
		excludedGPUs = append(excludedGPUs, config.Config.GPU.PrivilegedGPUs...)
	}

	if onlyAvailable {
		dbAvailableGPUs, err := gpuModel.NewWithExclusion(config.Config.GPU.ExcludedHosts, excludedGPUs).GetAllAvailable()
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
	return gpuModel.NewWithExclusion(config.Config.GPU.ExcludedHosts, excludedGPUs).GetAll()
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

func RepairGPUs() error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair gpus. details: %w", err)
	}

	// get all gpus that are attached to a vm
	attachedGPUs, err := gpuModel.New().GetAllLeased()
	if err != nil {
		return makeError(err)
	}

	// get all vms with an assigned gpu
	vmsWithGPU, err := vmModel.New().ListWithGPU()
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
			err = DetachGPU(vm2.ID)
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
			err = gpuModel.New().ClearLease(gpu.ID)
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
			err = DetachGPU(vm.ID)
			if err != nil {
				return makeError(err)
			}
		}
	}

	// find vms that have a gpu assigned, but cs is not setup to use it
	for _, vm := range vmsWithGPU {
		if !hasExtraConfig(&vm) {
			err := cs_service.AttachGPU(vm.GpuID, vm.ID)
			if err != nil {
				return makeError(err)
			}
		}
	}

	log.Println("successfully repaired gpus")
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

func CanStartOnHost(vmID, host string) (bool, string, error) {
	vm, err := vmModel.New().GetByID(vmID)
	if err != nil {
		return false, "", err
	}

	if vm == nil {
		return false, "", fmt.Errorf("vm %s not found", vmID)
	}

	if vm.Subsystems.CS.VM.ID == "" {
		return false, "VM not fully created", nil
	}

	canStart, reason, err := cs_service.CanStart(vm.Subsystems.CS.VM.ID, host, vm.Zone)
	if err != nil {
		return false, "", err
	}

	if !canStart {
		return false, reason, nil
	}

	return true, "", nil
}

func IsGpuHardwareAvailable(gpu *gpuModel.GPU) (bool, error) {
	cloudstackAttached, err := cs_service.IsGpuAttachedCS(gpu)
	if err != nil {
		return false, err
	}

	zone := config.Config.VM.GetZone(gpu.Zone)
	if zone == nil {
		return false, fmt.Errorf("zone %s not found", gpu.Zone)
	}

	correctState, _, err := cs_service.HostInCorrectState(gpu.Host, zone)
	if err != nil {
		return false, err
	}

	// a "bad attach" is when cloudstack reports it being bound, but the database says it's not
	// this usually means it is in use outside the scope of deploy
	badAttach := cloudstackAttached && !gpu.IsAttached()

	return !badAttach && correctState, nil
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

package cs_service

import (
	"fmt"
	vmModel "go-deploy/models/sys/vm"
	gpuModel "go-deploy/models/sys/vm/gpu"
	"go-deploy/pkg/conf"
	"go-deploy/service/vm_service/cs_service/helpers"
	"log"
	"strings"
)

func AttachGPU(gpuID, vmID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to attach gpu %s to cs vm %s. details: %w", gpuID, vmID, err)
	}

	vm, err := vmModel.New().GetByID(vmID)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		log.Println("vm", vmID, "not found when attaching gpu", gpuID, "to cs vm, assuming it was deleted")
		return nil
	}

	if vm.Subsystems.CS.VM.ID == "" {
		log.Println("vm", vmID, "has no cs vm id when attaching gpu", gpuID, "to cs vm, assuming it was deleted")
		return nil
	}

	zone := conf.Env.VM.GetZone(vm.Zone)
	if zone == nil {
		return makeError(fmt.Errorf("zone %s not found", vm.Zone))
	}

	client, err := helpers.WithCsClient(zone)
	if err != nil {
		return makeError(err)
	}

	gpu, err := gpuModel.New().GetByID(gpuID)
	if err != nil {
		return makeError(err)
	}

	requiredExtraConfig := helpers.CreateExtraConfig(gpu)
	currentExtraConfig := vm.Subsystems.CS.VM.ExtraConfig
	if requiredExtraConfig != currentExtraConfig {
		var status string
		status, err = client.GetVmStatus(vm.Subsystems.CS.VM.ID)
		if err != nil {
			return makeError(err)
		}

		if status == "Running" {
			err = client.DoVmCommand(vm.Subsystems.CS.VM.ID, nil, "stop")
			if err != nil {
				return makeError(err)
			}
		}

		vm.Subsystems.CS.VM.ExtraConfig = requiredExtraConfig

		err = client.UpdateVM(&vm.Subsystems.CS.VM)
		if err != nil {
			return makeError(err)
		}

		err = vmModel.New().UpdateSubsystemByName(vm.Name, "cs", "vm.extraConfig", vm.Subsystems.CS.VM.ExtraConfig)
		if err != nil {
			return makeError(err)
		}
	}

	// always start the vm after attaching gpu, to make sure the vm can be started on the host
	requiredHost, err := helpers.GetRequiredHost(gpuID)
	if err != nil {
		return makeError(err)
	}

	err = client.DoVmCommand(vm.Subsystems.CS.VM.ID, requiredHost, "start")
	if err != nil {
		return makeError(err)
	}

	return nil
}

func DetachGPU(vmID string, afterState string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to detach gpu from cs vm %s. details: %w", vmID, err)
	}

	vm, err := vmModel.New().GetByID(vmID)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		log.Println("vm", vmID, "not found for when detaching gpu in cs. assuming it was deleted")
		return nil
	}

	if vm.Subsystems.CS.VM.ID == "" {
		return nil
	}

	zone := conf.Env.VM.GetZone(vm.Zone)
	if zone == nil {
		return makeError(fmt.Errorf("zone %s not found", vm.Zone))
	}

	client, err := helpers.WithCsClient(zone)
	if err != nil {
		return makeError(err)
	}

	status, err := client.GetVmStatus(vm.Subsystems.CS.VM.ID)
	if err != nil {
		return makeError(err)
	}

	if status == "Running" {
		err = client.DoVmCommand(vm.Subsystems.CS.VM.ID, nil, "stop")
		if err != nil {
			return makeError(err)
		}
	}

	vm.Subsystems.CS.VM.ExtraConfig = ""

	err = client.UpdateVM(&vm.Subsystems.CS.VM)
	if err != nil {
		return makeError(err)
	}

	err = vmModel.New().UpdateSubsystemByName(vm.Name, "cs", "vm.extraConfig", vm.Subsystems.CS.VM.ExtraConfig)
	if err != nil {
		return makeError(err)
	}

	// turn it on if it was on
	if (status == "Running" && afterState == CsDetachGpuAfterStateRestore) || afterState == CsDetachGpuAfterStateOn {
		err = client.DoVmCommand(vm.Subsystems.CS.VM.ID, nil, "start")
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func IsGpuAttachedCS(gpu *gpuModel.GPU) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if gpu %s:%s is attached to any cs vm. details: %w", gpu.Host, gpu.Data.Bus, err)
	}

	zone := conf.Env.VM.GetZone(gpu.Zone)
	if zone == nil {
		return false, makeError(fmt.Errorf("zone %s not found", gpu.Zone))
	}

	client, err := helpers.WithCsClient(zone)
	if err != nil {
		return false, makeError(err)
	}

	params := client.CsClient.VirtualMachine.NewListVirtualMachinesParams()
	params.SetListall(true)

	vms, err := client.CsClient.VirtualMachine.ListVirtualMachines(params)
	if err != nil {
		return false, makeError(err)
	}

	for _, vm := range vms.VirtualMachines {
		if vm.Details != nil && vm.Hostname == gpu.Host {
			extraConfig, ok := vm.Details["extraconfig-1"]
			if ok {
				if strings.Contains(extraConfig, fmt.Sprintf("bus='0x%s'", gpu.Data.Bus)) {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

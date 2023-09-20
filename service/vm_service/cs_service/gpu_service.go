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

	client, err := helpers.New(&vm.Subsystems.CS, vm.Zone)
	if err != nil {
		return makeError(err)
	}

	if client.CS.VM.ID == "" {
		log.Println("vm", vmID, "has no cs vm id when attaching gpu", gpuID, "to cs vm, assuming it was deleted")
		return nil
	}

	gpu, err := gpuModel.New().GetByID(gpuID)
	if err != nil {
		return makeError(err)
	}

	requiredExtraConfig := CreateExtraConfig(gpu)
	currentExtraConfig := client.CS.VM.ExtraConfig
	if requiredExtraConfig != currentExtraConfig {
		var status string
		status, err = client.SsClient.GetVmStatus(client.CS.VM.ID)
		if err != nil {
			return makeError(err)
		}

		if status == "Running" {
			err = client.SsClient.DoVmCommand(client.CS.VM.ID, nil, "stop")
			if err != nil {
				return makeError(err)
			}
		}

		client.CS.VM.ExtraConfig = requiredExtraConfig

		err = client.SsClient.UpdateVM(&client.CS.VM)
		if err != nil {
			return makeError(err)
		}

		err = vmModel.New().UpdateSubsystemByName(vm.Name, "cs", "vm.extraConfig", client.CS.VM.ExtraConfig)
		if err != nil {
			return makeError(err)
		}
	}

	// always start the vm after attaching gpu, to make sure the vm can be started on the host
	requiredHost, err := GetRequiredHost(gpuID)
	if err != nil {
		return makeError(err)
	}

	err = client.SsClient.DoVmCommand(client.CS.VM.ID, requiredHost, "start")
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

	client, err := helpers.New(&vm.Subsystems.CS, vm.Zone)
	if err != nil {
		return makeError(err)
	}

	if !client.CS.VM.Created() {
		log.Println("csVM was not created for vm", vmID, "when detaching gpu in cs. assuming it was deleted or not created yet")
		return nil
	}

	status, err := client.SsClient.GetVmStatus(client.CS.VM.ID)
	if err != nil {
		return makeError(err)
	}

	if status == "Running" {
		err = client.SsClient.DoVmCommand(client.CS.VM.ID, nil, "stop")
		if err != nil {
			return makeError(err)
		}
	}

	client.CS.VM.ExtraConfig = ""

	err = client.SsClient.UpdateVM(&client.CS.VM)
	if err != nil {
		return makeError(err)
	}

	err = vmModel.New().UpdateSubsystemByName(vm.Name, "cs", "vm.extraConfig", client.CS.VM.ExtraConfig)
	if err != nil {
		return makeError(err)
	}

	// turn it on if it was on
	if (status == "Running" && afterState == CsDetachGpuAfterStateRestore) || afterState == CsDetachGpuAfterStateOn {
		err = client.SsClient.DoVmCommand(client.CS.VM.ID, nil, "start")
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

	client, err := helpers.New(nil, gpu.Zone)
	if err != nil {
		return false, makeError(err)
	}

	// this should be exposed through the subsystem api, but im too lazy to do it now
	params := client.SsClient.CsClient.VirtualMachine.NewListVirtualMachinesParams()
	params.SetListall(true)

	vms, err := client.SsClient.CsClient.VirtualMachine.ListVirtualMachines(params)
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

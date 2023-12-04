package cs_service

import (
	"errors"
	"fmt"
	gpuModel "go-deploy/models/sys/gpu"
	"go-deploy/pkg/subsystems/cs/commands"
	"go-deploy/service"
	"go-deploy/service/resources"
	"go-deploy/service/vm_service/base"
	"log"
	"strings"
)

func AttachGPU(gpuID, vmID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to attach gpu %s to cs vm %s. details: %w", gpuID, vmID, err)
	}

	context, err := NewContext(vmID)
	if err != nil {
		return makeError(err)
	}

	if service.NotCreated(&context.VM.Subsystems.CS.VM) {
		log.Println("vm", vmID, "has no cs vm id when attaching gpu", gpuID, "to cs vm, assuming it was deleted")
		return nil
	}

	gpu, err := gpuModel.New().GetByID(gpuID)
	if err != nil {
		return makeError(err)
	}

	requiredExtraConfig := CreateExtraConfig(gpu)
	currentExtraConfig := context.VM.Subsystems.CS.VM.ExtraConfig
	if requiredExtraConfig != currentExtraConfig {
		var status string
		status, err = context.Client.GetVmStatus(context.VM.Subsystems.CS.VM.ID)
		if err != nil {
			return makeError(err)
		}

		if status == "Running" {
			err = context.Client.DoVmCommand(context.VM.Subsystems.CS.VM.ID, nil, commands.Stop)
			if err != nil {
				return makeError(err)
			}
		}

		context.VM.Subsystems.CS.VM.ExtraConfig = requiredExtraConfig

		err = resources.SsUpdater(context.Client.UpdateVM).
			WithPublic(&context.VM.Subsystems.CS.VM).
			WithDbFunc(dbFunc(vmID, "vm")).
			Exec()

		if err != nil {
			return makeError(err)
		}

		requiredHost, err := GetRequiredHost(gpuID)
		if err != nil {
			return makeError(err)
		}

		err = context.Client.DoVmCommand(context.VM.Subsystems.CS.VM.ID, requiredHost, commands.Start)
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func DetachGPU(vmID string, afterState string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to detach gpu from cs vm %s. details: %w", vmID, err)
	}

	context, err := NewContext(vmID)
	if err != nil {
		if errors.Is(err, base.VmDeletedErr) {
			log.Println("vm", vmID, "not found when detaching gpu in cs. assuming it was deleted")
			return nil
		}

		return makeError(err)
	}

	if service.NotCreated(&context.VM.Subsystems.CS.VM) {
		log.Println("csVM was not created for vm", vmID, "when detaching gpu in cs. assuming it was deleted or not created yet")
		return nil
	}

	status, err := context.Client.GetVmStatus(context.VM.Subsystems.CS.VM.ID)
	if err != nil {
		return makeError(err)
	}

	if status == "Running" {
		err = context.Client.DoVmCommand(context.VM.Subsystems.CS.VM.ID, nil, commands.Stop)
		if err != nil {
			return makeError(err)
		}
	}

	context.VM.Subsystems.CS.VM.ExtraConfig = ""

	err = resources.SsUpdater(context.Client.UpdateVM).
		WithPublic(&context.VM.Subsystems.CS.VM).
		WithDbFunc(dbFunc(vmID, "vm")).
		Exec()

	if err != nil {
		return makeError(err)
	}

	// turn it on if it was on
	if (status == "Running" && afterState == CsDetachGpuAfterStateRestore) || afterState == CsDetachGpuAfterStateOn {
		err = context.Client.DoVmCommand(context.VM.Subsystems.CS.VM.ID, nil, "start")
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

	context, err := NewContextWithoutVM(gpu.Zone)
	if err != nil {
		return false, makeError(err)
	}

	// this should be exposed through the subsystem api, but im too lazy to do it now
	params := context.Client.CsClient.VirtualMachine.NewListVirtualMachinesParams()
	params.SetListall(true)

	vms, err := context.Client.CsClient.VirtualMachine.ListVirtualMachines(params)
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

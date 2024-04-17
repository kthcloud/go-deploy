package cs_service

import (
	"errors"
	"fmt"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/gpu_repo"
	"go-deploy/pkg/log"
	"go-deploy/pkg/subsystems"
	"go-deploy/pkg/subsystems/cs/commands"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/resources"
	"strings"
)

// AttachGPU attaches a GPU to a VM.
// It updates the VM's extra config and restarts the VM.
//
// If the VM is not running when attaching the GPU, it will not be started.
func (c *Client) AttachGPU(vmID, gpuID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to attach gpu %s to cs vm %s. details: %w", gpuID, vmID, err)
	}

	vm, csc, _, err := c.Get(OptsNoGenerator(vmID))
	if err != nil {
		if errors.Is(err, sErrors.VmNotFoundErr) {
			return nil
		}

		return makeError(err)
	}

	if subsystems.NotCreated(&vm.Subsystems.CS.VM) {
		log.Println("VM", vmID, "has no cs vm id when attaching gpu", gpuID, ", assuming it was deleted")
		return nil
	}

	gpu, err := gpu_repo.New().GetByID(gpuID)
	if err != nil {
		return makeError(err)
	}

	requiredExtraConfig := CreateExtraConfig(gpu)
	currentExtraConfig := vm.Subsystems.CS.VM.ExtraConfig
	if requiredExtraConfig != currentExtraConfig {
		var status string
		status, err = csc.GetVmStatus(vm.Subsystems.CS.VM.ID)
		if err != nil {
			return makeError(err)
		}

		shouldStartAfter := false
		if status == "Running" {
			shouldStartAfter = true
			err = csc.DoVmCommand(vm.Subsystems.CS.VM.ID, nil, commands.Stop)
			if err != nil {
				return makeError(err)
			}
		}

		vm.Subsystems.CS.VM.ExtraConfig = requiredExtraConfig

		err = resources.SsUpdater(csc.UpdateVM).
			WithPublic(&vm.Subsystems.CS.VM).
			WithDbFunc(dbFunc(vmID, "vm")).
			Exec()

		if err != nil {
			return makeError(err)
		}

		requiredHost, err := c.GetRequiredHost(gpuID)
		if err != nil {
			return makeError(err)
		}

		if shouldStartAfter {
			err = csc.DoVmCommand(vm.Subsystems.CS.VM.ID, requiredHost, commands.Start)
			if err != nil {
				return makeError(err)
			}
		}

	}

	return nil
}

// DetachGPU detaches a GPU from a VM.
// It updates the VM's extra config and restarts the VM.
//
// If the VM is not running when detaching the GPU, it will not be started.
func (c *Client) DetachGPU(vmID string, afterState string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to detach gpu from cs vm %s. details: %w", vmID, err)
	}

	vm, csc, _, err := c.Get(OptsNoGenerator(vmID))
	if err != nil {
		if errors.Is(err, sErrors.VmNotFoundErr) {
			return nil
		}

		return makeError(err)
	}

	if subsystems.NotCreated(&vm.Subsystems.CS.VM) {
		log.Println("CSVM was not created for vm", vmID, "when detaching gpu in cs. Assuming it was deleted or not created yet")
		return nil
	}

	status, err := csc.GetVmStatus(vm.Subsystems.CS.VM.ID)
	if err != nil {
		return makeError(err)
	}

	if status == "Running" {
		err = csc.DoVmCommand(vm.Subsystems.CS.VM.ID, nil, commands.Stop)
		if err != nil {
			return makeError(err)
		}
	}

	vm.Subsystems.CS.VM.ExtraConfig = ""

	err = resources.SsUpdater(csc.UpdateVM).
		WithPublic(&vm.Subsystems.CS.VM).
		WithDbFunc(dbFunc(vmID, "vm")).
		Exec()

	if err != nil {
		return makeError(err)
	}

	// turn it on if it was on
	if (status == "Running" && afterState == CsDetachGpuAfterStateRestore) || afterState == CsDetachGpuAfterStateOn {
		err = csc.DoVmCommand(vm.Subsystems.CS.VM.ID, nil, "start")
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

// IsGpuAttached checks if a GPU is attached to any CS VM.
// It does this by inspecting the extra config of all VMs (including those not managed by go-deploy).
func (c *Client) IsGpuAttached(id string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if gpu %s is attached to any cs vm. details: %w", id, err)
	}

	gpu, err := c.GPU(id, nil)
	if err != nil {
		return false, makeError(err)
	}

	if gpu == nil {
		return false, nil
	}

	zone := config.Config.VM.GetLegacyZone(gpu.Zone)
	if zone == nil {
		return false, makeError(sErrors.ZoneNotFoundErr)
	}

	_, csc, _, err := c.Get(OptsOnlyClient(zone))
	if err != nil {
		return false, makeError(err)
	}

	// this should be exposed through the subsystem api, but im too lazy to do it now
	params := csc.CsClient.VirtualMachine.NewListVirtualMachinesParams()
	params.SetListall(true)

	vms, err := csc.CsClient.VirtualMachine.ListVirtualMachines(params)
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

// GetRequiredHost returns the host a GPU is required to be on.
// This is needed because the GPU is attached using GPU passthrough, which requires the VM to be on the same host.
func (c *Client) GetRequiredHost(gpuID string) (*string, error) {
	gpu, err := c.GPU(gpuID, nil)
	if err != nil {
		return nil, err
	}

	if gpu.Host == "" {
		return nil, fmt.Errorf("no host found for gpu_repo %s", gpu.ID)
	}

	return &gpu.Host, nil
}

package cs_service

import (
	"fmt"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/subsystems/cs/commands"
	csModels "go-deploy/pkg/subsystems/cs/models"
	"go-deploy/service/vm_service/cs_service/helpers"
	"log"
)

func CreateSnapshotCS(vmID, name string, userCreated bool) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create snapshot for cs vm %s. details: %w", vmID, err)
	}

	vm, err := vmModel.New().GetByID(vmID)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		log.Println("vm", vmID, "not found for when creating snapshot in cs. assuming it was deleted")
		return nil
	}

	client, err := helpers.New(&vm.Subsystems.CS, vm.Zone)
	if err != nil {
		return makeError(err)
	}

	snapshotMap := vm.Subsystems.CS.SnapshotMap
	if snapshotMap == nil {
		snapshotMap = map[string]csModels.SnapshotPublic{}
	}

	if _, ok := snapshotMap[name]; ok {
		log.Println("snapshot", name, "already exists for vm", vmID)
		return nil
	}

	vmStatus, err := client.SsClient.GetVmStatus(vm.Subsystems.CS.VM.ID)
	if err != nil {
		return makeError(err)
	}

	if vmStatus != "Running" {
		return fmt.Errorf("cs vm %s is not running", vm.Subsystems.CS.VM.ID)
	}

	var description string
	if userCreated {
		description = "go-deploy user"
	} else {
		description = "go-deploy system"
	}

	public := &csModels.SnapshotPublic{
		Name:        name,
		VmID:        vm.Subsystems.CS.VM.ID,
		Description: description,
	}

	var gpuID *string
	if HasExtraConfig(vm) {
		gpuID = &vm.GpuID
		err := DetachGPU(vm.ID, CsDetachGpuAfterStateOn)
		if err != nil {
			return makeError(err)
		}
	}

	// make sure vm is on
	err = client.SsClient.DoVmCommand(vm.Subsystems.CS.VM.ID, nil, commands.Start)
	if err != nil {
		return makeError(err)
	}

	snapshotID, err := client.SsClient.CreateSnapshot(public)
	if err != nil {
		return makeError(err)
	}

	if gpuID != nil {
		err := AttachGPU(*gpuID, vmID)
		if err != nil {
			return makeError(err)
		}
	}

	log.Println("created snapshot", snapshotID, "for vm", vmID)

	return nil
}

func ApplySnapshotCS(vmID, snapshotID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to apply snapshot %s for vm %s. details: %w", snapshotID, vmID, err)
	}

	vm, err := vmModel.New().GetByID(vmID)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		log.Println("vm", vmID, "not found for when applying snapshot in cs. assuming it was deleted")
		return nil
	}

	client, err := helpers.New(&vm.Subsystems.CS, vm.Zone)
	if err != nil {
		return makeError(err)
	}

	snapshotMap := vm.Subsystems.CS.SnapshotMap
	if snapshotMap == nil {
		snapshotMap = map[string]csModels.SnapshotPublic{}
	}

	snapshot, ok := snapshotMap[snapshotID]
	if !ok {
		return fmt.Errorf("snapshot %s not found", snapshotID)
	}

	if snapshot.State != "Ready" {
		return fmt.Errorf("snapshot %s is not ready", snapshotID)
	}

	var gpuID *string
	if HasExtraConfig(vm) {
		gpuID = &vm.GpuID
		err := DetachGPU(vm.ID, CsDetachGpuAfterStateOn)
		if err != nil {
			return makeError(err)
		}
	}

	// make sure vm is on
	err = client.SsClient.DoVmCommand(vm.Subsystems.CS.VM.ID, nil, commands.Start)
	if err != nil {
		return makeError(err)
	}

	err = client.SsClient.ApplySnapshot(&snapshot)
	if err != nil {
		return makeError(err)
	}

	if gpuID != nil {
		err := AttachGPU(*gpuID, vmID)
		if err != nil {
			return makeError(err)
		}
	}

	log.Println("applied snapshot", snapshotID, "for vm", vmID)

	return nil
}

package cs_service

import (
	"fmt"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/conf"
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

	zone := conf.Env.VM.GetZone(vm.Zone)
	if zone == nil {
		return makeError(fmt.Errorf("zone %s not found", vm.Zone))
	}

	client, err := helpers.WithCsClient(zone)
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

	vmStatus, err := client.GetVmStatus(vm.Subsystems.CS.VM.ID)
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
	if helpers.HasExtraConfig(vm) {
		gpuID = &vm.GpuID
		err := DetachGPU(vm.ID, CsDetachGpuAfterStateOn)
		if err != nil {
			return makeError(err)
		}
	}

	// make sure vm is on
	err = client.DoVmCommand(vm.Subsystems.CS.VM.ID, nil, "start")
	if err != nil {
		return makeError(err)
	}

	snapshotID, err := client.CreateSnapshot(public)
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

	zone := conf.Env.VM.GetZone(vm.Zone)
	if zone == nil {
		return makeError(fmt.Errorf("zone %s not found", vm.Zone))
	}

	client, err := helpers.WithCsClient(zone)
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
	if helpers.HasExtraConfig(vm) {
		gpuID = &vm.GpuID
		err := DetachGPU(vm.ID, CsDetachGpuAfterStateOn)
		if err != nil {
			return makeError(err)
		}
	}

	// make sure vm is on
	err = client.DoVmCommand(vm.Subsystems.CS.VM.ID, nil, "start")
	if err != nil {
		return makeError(err)
	}

	err = client.ApplySnapshot(&snapshot)
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

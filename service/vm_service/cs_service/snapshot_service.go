package cs_service

import (
	"errors"
	"fmt"
	"go-deploy/pkg/subsystems/cs/commands"
	csModels "go-deploy/pkg/subsystems/cs/models"
	"go-deploy/service"
	"go-deploy/service/vm_service/base"
	"log"
)

func CreateSnapshotCS(vmID, name string, userCreated bool) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create snapshot for cs vm %s. details: %w", vmID, err)
	}

	context, err := NewContext(vmID)
	if err != nil {
		if errors.Is(err, base.VmDeletedErr) {
			log.Println("vm", vmID, "not found when creating snapshot in cs. assuming it was deleted")
			return nil
		}

		return makeError(err)
	}

	if snapshot := context.VM.Subsystems.CS.GetSnapshot(name); service.Created(snapshot) {
		log.Println("snapshot", name, "already exists for vm", vmID)
		return nil
	}

	vmStatus, err := context.Client.GetVmStatus(context.VM.Subsystems.CS.VM.ID)
	if err != nil {
		return makeError(err)
	}

	if vmStatus != "Running" {
		return fmt.Errorf("cs vm %s is not running", context.VM.Subsystems.CS.VM.ID)
	}

	var description string
	if userCreated {
		description = "go-deploy user"
	} else {
		description = "go-deploy system"
	}

	public := &csModels.SnapshotPublic{
		Name:        name,
		VmID:        context.VM.Subsystems.CS.VM.ID,
		Description: description,
	}

	var gpuID *string
	if HasExtraConfig(context.VM) {
		gpuID = &context.VM.GpuID
		err := DetachGPU(context.VM.ID, CsDetachGpuAfterStateOn)
		if err != nil {
			return makeError(err)
		}
	}

	// make sure vm is on
	err = context.Client.DoVmCommand(context.VM.Subsystems.CS.VM.ID, nil, commands.Start)
	if err != nil {
		return makeError(err)
	}

	snapshotID, err := context.Client.CreateSnapshot(public)
	if err != nil {
		return makeError(err)
	}

	if gpuID != nil {
		err = AttachGPU(*gpuID, vmID)
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

	context, err := NewContext(vmID)
	if err != nil {
		if errors.Is(err, base.VmDeletedErr) {
			log.Println("vm", vmID, "not found when applying snapshot in cs. assuming it was deleted")
			return nil
		}

		return makeError(err)
	}

	snapshot := context.VM.Subsystems.CS.GetSnapshot(snapshotID)
	if service.NotCreated(snapshot) {
		return makeError(fmt.Errorf("snapshot %s not found", snapshotID))
	}

	if snapshot.State != "Ready" {
		return makeError(fmt.Errorf("snapshot %s is not ready", snapshotID))
	}

	var gpuID *string
	if HasExtraConfig(context.VM) {
		gpuID = &context.VM.GpuID
		err := DetachGPU(vmID, CsDetachGpuAfterStateOn)
		if err != nil {
			return makeError(err)
		}
	}

	// make sure vm is on
	err = context.Client.DoVmCommand(context.VM.Subsystems.CS.VM.ID, nil, commands.Start)
	if err != nil {
		return makeError(err)
	}

	err = context.Client.ApplySnapshot(snapshot)
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

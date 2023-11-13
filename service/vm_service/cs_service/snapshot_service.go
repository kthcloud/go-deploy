package cs_service

import (
	"errors"
	"fmt"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/subsystems/cs/commands"
	csModels "go-deploy/pkg/subsystems/cs/models"
	"go-deploy/service"
	"go-deploy/service/vm_service/base"
	"log"
)

var (
	AlreadyExistsErr = fmt.Errorf("already exists")
)

func CreateSnapshot(vmID string, params *vmModel.CreateSnapshotParams) error {
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

	if snapshot := context.VM.Subsystems.CS.GetSnapshotByName(params.Name); service.Created(snapshot) && !params.Overwrite {
		return AlreadyExistsErr
	}

	vmStatus, err := context.Client.GetVmStatus(context.VM.Subsystems.CS.VM.ID)
	if err != nil {
		return makeError(err)
	}

	if vmStatus != "Running" {
		return fmt.Errorf("cs vm %s is not running", context.VM.Subsystems.CS.VM.ID)
	}

	var description string
	if params.UserCreated {
		description = "go-deploy user"
	} else {
		description = "go-deploy system"
	}

	public := &csModels.SnapshotPublic{
		Name:        params.Name,
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

	// delete every other snapshot with the same name that is older
	snapshots, err := context.Client.ReadAllSnapshots(context.VM.Subsystems.CS.VM.ID)
	if err != nil {
		return makeError(err)
	}

	for _, snapshot := range snapshots {
		if snapshot.Name == params.Name && snapshot.ID != snapshotID {
			err = context.Client.DeleteSnapshot(snapshot.ID)
			if err != nil {
				return makeError(err)
			}
		}
	}

	log.Println("created snapshot", snapshotID, "for vm", vmID)

	return nil
}

func DeleteSnapshot(vmID, snapshotID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete snapshot %s for vm %s. details: %w", snapshotID, vmID, err)
	}

	context, err := NewContext(vmID)
	if err != nil {
		if errors.Is(err, base.VmDeletedErr) {
			log.Println("vm", vmID, "not found when deleting snapshot in cs. assuming it was deleted")
			return nil
		}

		return makeError(err)
	}

	err = context.Client.DeleteSnapshot(snapshotID)
	if err != nil {
		return makeError(err)
	}

	log.Println("deleted snapshot", snapshotID, "for vm", vmID)

	return nil
}

func ApplySnapshot(vmID, snapshotID string) error {
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

	snapshot := context.VM.Subsystems.CS.GetSnapshotByID(snapshotID)
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

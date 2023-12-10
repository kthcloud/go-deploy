package cs_service

import (
	"errors"
	"fmt"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/subsystems/cs/commands"
	csModels "go-deploy/pkg/subsystems/cs/models"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/vm_service/client"
	"log"
)

var (
	AlreadyExistsErr = fmt.Errorf("already exists")
	BadStateErr      = fmt.Errorf("bad state")
)

func makeBadStateErr(state string) error {
	return fmt.Errorf("%w: %s", BadStateErr, state)
}

func (c *Client) CreateSnapshot(vmID string, params *vmModel.CreateSnapshotParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create snapshot for cs vm %s. details: %w", vmID, err)
	}

	vm, csc, _, err := c.Get(client.OptsNoGenerator(vmID))
	if err != nil {
		if errors.Is(err, sErrors.VmNotFoundErr) {
			return nil
		}

		return makeError(err)
	}

	if snapshot := vm.Subsystems.CS.GetSnapshotByName(params.Name); service.Created(snapshot) && !params.Overwrite {
		return AlreadyExistsErr
	}

	// Make sure vm is on
	vmStatus, err := csc.GetVmStatus(vm.Subsystems.CS.VM.ID)
	if err != nil {
		return makeError(err)
	}

	if vmStatus != "Running" {
		return makeBadStateErr(fmt.Sprintf("cs vm %s is not in running state: %s", vm.Subsystems.CS.VM.ID, vmStatus))
	}

	var description string
	if params.UserCreated {
		description = "go-deploy user"
	} else {
		description = "go-deploy system"
	}

	public := &csModels.SnapshotPublic{
		Name:        params.Name,
		VmID:        vm.Subsystems.CS.VM.ID,
		Description: description,
	}

	if HasExtraConfig(vm) {
		return makeBadStateErr("vm has extra config (probably a gpu attached)")
	}

	snapshotID, err := csc.CreateSnapshot(public)
	if err != nil {
		return makeError(err)
	}

	if snapshotID == "" {
		// If no error was received, and no snapshot ID was received, then the snapshot was gracefully not created
		// So we don't return any error here
		return nil
	}

	if !params.UserCreated {
		// Fetch to see what state the snapshot is in, in order to delete the bad ones
		snapshot, err := csc.ReadSnapshot(snapshotID)
		if err != nil {
			_ = csc.DeleteSnapshot(snapshotID)
			return makeError(err)
		}

		if snapshot != nil && snapshot.State == "Error" {
			_ = csc.DeleteSnapshot(snapshotID)
			return makeBadStateErr(fmt.Sprintf("snapshot got state: %s", snapshot.State))
		}
	}
	log.Println("created snapshot", snapshotID, "for vm", vmID)

	// Delete every other snapshot with the same name that is older
	snapshots, err := csc.ReadAllSnapshots(vm.Subsystems.CS.VM.ID)
	if err != nil {
		return makeError(err)
	}

	for _, snapshot := range snapshots {
		if snapshot.Name == params.Name && snapshot.ID != snapshotID {
			err = csc.DeleteSnapshot(snapshot.ID)
			if err != nil {
				return makeError(err)
			}

			log.Println("deleted old snapshot", snapshot.ID, "for vm", vmID)
		}
	}

	return nil
}

func (c *Client) DeleteSnapshot(vmID, snapshotID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete snapshot %s for vm %s. details: %w", snapshotID, vmID, err)
	}

	_, csc, _, err := c.Get(client.OptsOnlyClient())
	if err != nil {
		if errors.Is(err, sErrors.VmNotFoundErr) {
			return nil
		}

		return makeError(err)
	}

	err = csc.DeleteSnapshot(snapshotID)
	if err != nil {
		return makeError(err)
	}

	log.Println("deleted snapshot", snapshotID, "for vm", vmID)

	return nil
}

func (c *Client) ApplySnapshot(vmID, snapshotID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to apply snapshot %s for vm %s. details: %w", snapshotID, vmID, err)
	}

	vm, csc, _, err := c.Get(client.OptsNoGenerator(vmID))
	if err != nil {
		if errors.Is(err, sErrors.VmNotFoundErr) {
			return nil
		}

		return makeError(err)
	}

	snapshot := vm.Subsystems.CS.GetSnapshotByID(snapshotID)
	if service.NotCreated(snapshot) {
		return makeError(fmt.Errorf("snapshot %s not found", snapshotID))
	}

	if snapshot.State != "Ready" {
		return makeError(fmt.Errorf("snapshot %s is not ready", snapshotID))
	}

	var gpuID *string
	if HasExtraConfig(vm) {
		gpuID = vm.GetGpuID()
		err = c.DetachGPU(vmID, CsDetachGpuAfterStateOn)
		if err != nil {
			return makeError(err)
		}
	}

	// make sure vm is on
	err = csc.DoVmCommand(vm.Subsystems.CS.VM.ID, nil, commands.Start)
	if err != nil {
		return makeError(err)
	}

	err = csc.ApplySnapshot(snapshot)
	if err != nil {
		return makeError(err)
	}

	if gpuID != nil {
		err = c.AttachGPU(vmID, *gpuID)
		if err != nil {
			return makeError(err)
		}
	}

	log.Println("applied snapshot", snapshotID, "for vm", vmID)

	return nil
}

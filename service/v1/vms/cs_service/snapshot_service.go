package cs_service

import (
	"errors"
	"fmt"
	"go-deploy/models/model"
	"go-deploy/pkg/db/resources/gpu_repo"
	"go-deploy/pkg/log"
	"go-deploy/pkg/subsystems"
	"go-deploy/pkg/subsystems/cs/commands"
	csModels "go-deploy/pkg/subsystems/cs/models"
	sErrors "go-deploy/service/errors"
)

// makeBadStateErr creates an error with the BadStateErr type.
func makeBadStateErr(state string) error {
	return fmt.Errorf("%w %s", sErrors.BadStateErr, state)
}

// CreateSnapshot creates a snapshot for a VM.
//
// If the VM is not running, it will return a BadStateErr.
func (c *Client) CreateSnapshot(vmID string, params *model.CreateSnapshotParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create snapshot for cs vm %s. details: %w", vmID, err)
	}

	vm, csc, _, err := c.Get(OptsNoGenerator(vmID))
	if err != nil {
		if errors.Is(err, sErrors.VmNotFoundErr) {
			return nil
		}

		return makeError(err)
	}

	if snapshot := vm.Subsystems.CS.GetSnapshotByName(params.Name); subsystems.Created(snapshot) && !params.Overwrite {
		return sErrors.SnapshotAlreadyExistsErr
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
		return makeBadStateErr("vm has extra config (probably a gpu_repo attached)")
	}

	snapshot, err := csc.CreateSnapshot(public)
	if err != nil {
		return makeError(err)
	}

	if snapshot == nil {
		// If no error was received, and no snapshot was received, then the snapshot was gracefully not created
		// So we don't return any error here
		return nil
	}

	if !params.UserCreated {
		if snapshot.State == "Error" {
			_ = csc.DeleteSnapshot(snapshot.ID)
			return makeBadStateErr(fmt.Sprintf("snapshot got state: %s", snapshot.State))
		}
	}
	log.Println("created snapshot", snapshot.ID, "for vm", vmID)

	// Delete every other snapshot with the same name that is older
	snapshots, err := csc.ReadAllSnapshots(vm.Subsystems.CS.VM.ID)
	if err != nil {
		return makeError(err)
	}

	for _, s := range snapshots {
		if s.Name == params.Name && s.ID != snapshot.ID {
			err = csc.DeleteSnapshot(s.ID)
			if err != nil {
				return makeError(err)
			}

			log.Println("deleted old snapshot", s.ID, "for vm", vmID)
		}
	}

	return nil
}

// DeleteSnapshot deletes a snapshot for a VM.
func (c *Client) DeleteSnapshot(vmID, snapshotID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete snapshot %s for vm %s. details: %w", snapshotID, vmID, err)
	}

	_, csc, _, err := c.Get(OptsNoGenerator(vmID))
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

// ApplySnapshot applies a snapshot for a VM.
// If the VM is not running, it will be started.
//
// If the VM has a GPU attached, it will be detached and re-attached after the snapshot is applied.
func (c *Client) ApplySnapshot(vmID, snapshotID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to apply snapshot %s for vm %s. details: %w", snapshotID, vmID, err)
	}

	vm, csc, _, err := c.Get(OptsNoGenerator(vmID))
	if err != nil {
		if errors.Is(err, sErrors.VmNotFoundErr) {
			return nil
		}

		return makeError(err)
	}

	snapshot := vm.Subsystems.CS.GetSnapshotByID(snapshotID)
	if subsystems.NotCreated(snapshot) {
		return makeError(sErrors.SnapshotNotFoundErr)
	}

	if snapshot.State != "Ready" {
		return makeError(fmt.Errorf("snapshot %s is not ready", snapshotID))
	}

	var gpuID *string
	if HasExtraConfig(vm) {
		gpuID, err = gpu_repo.New().WithVM(vm.ID).GetID()
		err = c.DetachGPU(vmID, CsDetachGpuAfterStateOn)
		if err != nil {
			return makeError(err)
		}
	}

	// Make sure vm is on
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

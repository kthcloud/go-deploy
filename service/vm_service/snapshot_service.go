package vm_service

import (
	"fmt"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/service/vm_service/internal_service"
	"log"
)

func GetSnapshotsByVM(vmID string) ([]vmModel.Snapshot, error) {
	vm, err := vmModel.GetByID(vmID)
	if err != nil {
		return nil, err
	}

	if vm == nil {
		return nil, nil
	}

	snapshots := make([]vmModel.Snapshot, 0)
	for _, snapshot := range vm.Subsystems.CS.SnapshotMap {
		snapshots = append(snapshots, vmModel.Snapshot{
			ID:         snapshot.ID,
			VmID:       vmID,
			Name:       snapshot.Name,
			ParentName: snapshot.ParentName,
			CreatedAt:  snapshot.CreatedAt,
			State:      snapshot.State,
			Current:    snapshot.Current,
		})
	}

	return snapshots, nil
}

func CreateSnapshot(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create snapshot for vm %s. details: %s", id, err)
	}

	vm, err := vmModel.GetByID(id)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		return fmt.Errorf("vm %s not found", id)
	}

	if !vm.Ready() {
		return fmt.Errorf("vm %s not ready", id)
	}

	started, reason, err := StartActivity(vm.ID, vmModel.ActivityCreatingSnapshot)
	if err != nil {
		return makeError(err)
	}

	if !started {
		return fmt.Errorf("failed to create snapshot for vm %s. details: %s", id, reason)
	}

	defer func() {
		err = vmModel.RemoveActivity(vm.ID, vmModel.ActivityCreatingSnapshot)
		if err != nil {
			log.Println("failed to remove activity", vmModel.ActivityCreatingSnapshot, "from vm", vm.Name, "details:", err)
		}
	}()

	err = internal_service.CreateSnapshotCS(vm.ID)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func ApplySnapshot(id, snapshotID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to apply snapshot %s to vm %s. details: %s", snapshotID, id, err)
	}

	log.Println("applying snapshot", snapshotID, "to vm", id)
	started, reason, err := StartActivity(id, vmModel.ActivityApplyingSnapshot)
	if err != nil {
		return makeError(err)
	}

	if !started {
		return fmt.Errorf("failed to apply snapshot %s to vm %s. details: %s", snapshotID, id, reason)
	}

	defer func() {
		_ = vmModel.RemoveActivity(id, vmModel.ActivityApplyingSnapshot)
	}()

	err = internal_service.ApplySnapshotCS(id, snapshotID)
	if err != nil {
		return makeError(err)
	}

	return nil
}

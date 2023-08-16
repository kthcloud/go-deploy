package vm_service

import (
	"fmt"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/service/vm_service/internal_service"
	"log"
	"time"
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
	log.Println("applying snapshot", snapshotID, "to vm", id)
	_, _, _ = StartActivity(id, vmModel.ActivityApplyingSnapshot)

	time.Sleep(10 * time.Second)

	_ = vmModel.RemoveActivity(id, vmModel.ActivityApplyingSnapshot)

	return nil
}

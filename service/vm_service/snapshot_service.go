package vm_service

import (
	"fmt"
	roleModel "go-deploy/models/sys/enviroment/role"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/service/vm_service/internal_service"
	"go-deploy/utils"
	"log"
	"sort"
)

func GetSnapshotsByVM(vmID string) ([]vmModel.Snapshot, error) {
	vm, err := vmModel.New().GetByID(vmID)
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

	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].CreatedAt.Before(snapshots[j].CreatedAt)
	})

	return snapshots, nil
}

func GetSnapshotByName(vmID, snapshotName string) (*vmModel.Snapshot, error) {
	vm, err := vmModel.New().GetByID(vmID)
	if err != nil {
		return nil, err
	}

	if vm == nil {
		return nil, nil
	}

	snapshot, ok := vm.Subsystems.CS.SnapshotMap[snapshotName]
	if !ok {
		return nil, nil
	}

	return &vmModel.Snapshot{
		ID:         snapshot.ID,
		VmID:       vmID,
		Name:       snapshot.Name,
		ParentName: snapshot.ParentName,
		CreatedAt:  snapshot.CreatedAt,
		State:      snapshot.State,
		Current:    snapshot.Current,
	}, nil
}

func CreateSnapshot(vmID, name string, userCreated bool) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create snapshot for vm %s. details: %w", vmID, err)
	}

	vm, err := vmModel.New().GetByID(vmID)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		log.Println("vm", vmID, "not found when creating snapshot. assuming it was deleted")
		return nil
	}

	if !vm.Ready() {
		return fmt.Errorf("vm %s not ready", vmID)
	}

	started, reason, err := StartActivity(vm.ID, vmModel.ActivityCreatingSnapshot)
	if err != nil {
		return makeError(err)
	}

	if !started {
		return fmt.Errorf("failed to create snapshot for vm %s. details: %s", vmID, reason)
	}

	defer func() {
		err = vmModel.New().RemoveActivity(vm.ID, vmModel.ActivityCreatingSnapshot)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to remove activity %s from vm %s. details: %w", vmModel.ActivityCreatingSnapshot, vm.Name, err))
		}
	}()

	err = internal_service.CreateSnapshotCS(vm.ID, name, userCreated)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func ApplySnapshot(id, snapshotID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to apply snapshot %s to vm %s. details: %w", snapshotID, id, err)
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
		_ = vmModel.New().RemoveActivity(id, vmModel.ActivityApplyingSnapshot)
	}()

	err = internal_service.ApplySnapshotCS(id, snapshotID)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func CheckQuotaCreateSnapshot(userID string, quota *roleModel.Quotas) (bool, string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check quota. details: %w", err)
	}

	usage, err := GetUsageByUserID(userID)
	if err != nil {
		return false, "", makeError(err)
	}

	if usage.Snapshots >= quota.Snapshots {
		return false, fmt.Sprintf("Snapshot count quota exceeded. Current: %d, Quota: %d", usage.Snapshots, quota.Snapshots), nil
	}

	return true, "", nil
}

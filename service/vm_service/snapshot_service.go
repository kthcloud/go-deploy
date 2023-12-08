package vm_service

import (
	"fmt"
	"go-deploy/models/dto/body"
	roleModel "go-deploy/models/sys/role"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/service"
	"go-deploy/service/errors"
	"go-deploy/service/vm_service/cs_service"
	"log"
	"sort"
)

func ListSnapshotsByVM(vmID string) ([]vmModel.Snapshot, error) {
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

	snapshot := vm.Subsystems.CS.GetSnapshotByName(snapshotName)
	if snapshot == nil {
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

func GetSnapshotByID(vmID, snapshotID string) (*vmModel.Snapshot, error) {
	vm, err := vmModel.New().GetByID(vmID)
	if err != nil {
		return nil, err
	}

	if vm == nil {
		return nil, nil
	}

	snapshot := vm.Subsystems.CS.GetSnapshotByID(snapshotID)
	if snapshot == nil {
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

func CreateSystemSnapshot(vmID string, params *vmModel.CreateSnapshotParams) error {
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

	err = cs_service.CreateSnapshot(vm.ID, params)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func CreateUserSnapshot(vmID string, dtoCreateSnapshot *body.VmSnapshotCreate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create snapshot for vm %s. details: %w", vmID, err)
	}

	params := &vmModel.CreateSnapshotParams{}
	params.FromDTO(dtoCreateSnapshot)

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

	err = cs_service.CreateSnapshot(vm.ID, params)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func DeleteSnapshot(vmID, snapshotID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete snapshot %s from vm %s. details: %w", snapshotID, vmID, err)
	}

	vm, err := vmModel.New().GetByID(vmID)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		log.Println("vm", vmID, "not found when deleting snapshot", snapshotID, ". assuming it was deleted")
		return nil
	}

	if !vm.Ready() {
		return fmt.Errorf("vm %s not ready", vmID)
	}

	err = cs_service.DeleteSnapshot(vm.ID, snapshotID)
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
	err := cs_service.ApplySnapshot(id, snapshotID)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func CheckQuotaCreateSnapshot(userID string, quota *roleModel.Quotas, auth *service.AuthInfo) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check quota. details: %w", err)
	}

	if auth.IsAdmin {
		return nil
	}

	usage, err := GetUsageByUserID(userID)
	if err != nil {
		return makeError(err)
	}

	if usage.Snapshots >= quota.Snapshots {
		return errors.NewQuotaExceededError(fmt.Sprintf("Snapshot count quota exceeded. Current: %d, Quota: %d", usage.Snapshots, quota.Snapshots))
	}

	return nil
}

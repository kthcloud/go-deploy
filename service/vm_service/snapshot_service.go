package vm_service

import (
	"fmt"
	vmModels "go-deploy/models/sys/vm"
	"go-deploy/service"
	"go-deploy/service/vm_service/client"
	"go-deploy/service/vm_service/cs_service"
	"log"
	"sort"
)

func (c *Client) GetSnapshot(vmID string, id string, opts ...client.GetSnapshotOptions) (*vmModels.Snapshot, error) {
	_ = service.GetFirstOrDefault(opts)

	vm, err := c.Get(vmID)
	if err != nil {
		return nil, err
	}

	if vm == nil {
		return nil, nil
	}

	snapshot := vm.Subsystems.CS.GetSnapshotByID(id)
	if snapshot == nil {
		return nil, nil
	}

	return &vmModels.Snapshot{
		ID:         snapshot.ID,
		VmID:       vmID,
		Name:       snapshot.Name,
		ParentName: snapshot.ParentName,
		CreatedAt:  snapshot.CreatedAt,
		State:      snapshot.State,
		Current:    snapshot.Current,
	}, nil
}

func (c *Client) GetSnapshotByName(vmID string, name string, opts ...client.GetSnapshotOptions) (*vmModels.Snapshot, error) {
	_ = service.GetFirstOrDefault(opts)

	vm, err := c.Get(vmID)
	if err != nil {
		return nil, err
	}

	if vm == nil {
		return nil, nil
	}

	snapshot := vm.Subsystems.CS.GetSnapshotByName(name)
	if snapshot == nil {
		return nil, nil
	}

	return &vmModels.Snapshot{
		ID:         snapshot.ID,
		VmID:       vmID,
		Name:       snapshot.Name,
		ParentName: snapshot.ParentName,
		CreatedAt:  snapshot.CreatedAt,
		State:      snapshot.State,
		Current:    snapshot.Current,
	}, nil
}

func (c *Client) ListSnapshots(vmID string, opts ...client.ListSnapshotOptions) ([]vmModels.Snapshot, error) {
	_ = service.GetFirstOrDefault(opts)

	vm, err := c.Get(vmID)
	if err != nil {
		return nil, err
	}

	if vm == nil {
		return nil, nil
	}

	snapshots := make([]vmModels.Snapshot, 0)
	for _, snapshot := range vm.Subsystems.CS.SnapshotMap {
		snapshots = append(snapshots, vmModels.Snapshot{
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

func (c *Client) CreateSnapshot(vmID string, opts *client.CreateSnapshotOptions) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create snapshot for vm %s. details: %w", vmID, err)
	}

	vm, err := c.Get(vmID)
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
  
	params := &vmModels.CreateSnapshotParams{}
	if opts.System != nil {
		params = opts.System
	} else if opts.User != nil {
		params.FromDTO(opts.User)
	}

	if params.Name == "" {
		log.Println("no snapshot type specified when creating snapshot for vm", vmID, ". did you forget to specify the type?")
		return nil
	}

	err = cs_service.New(c.Cache).CreateSnapshot(vm.ID, params)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func (c *Client) DeleteSnapshot(vmID, snapshotID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete snapshot %s from vm %s. details: %w", snapshotID, vmID, err)
	}

	vm, err := c.Get(vmID)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		log.Println("vm", vmID, "not found when deleting snapshot", snapshotID, ". assuming it was deleted")
		return nil
	}

	if !vm.Ready() {
		return makeError(fmt.Errorf("vm not ready"))
	}

	err = cs_service.New(c.Cache).DeleteSnapshot(vm.ID, snapshotID)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func (c *Client) ApplySnapshot(id, snapshotID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to apply snapshot %s to vm %s. details: %w", snapshotID, id, err)
	}

	log.Println("applying snapshot", snapshotID, "to vm", id)
	err := cs_service.New(c.Cache).ApplySnapshot(id, snapshotID)
	if err != nil {
		return makeError(err)
	}

	return nil
}

package vms

import (
	"fmt"
	"go-deploy/models/model"
	"go-deploy/pkg/log"
	"go-deploy/service/utils"
	"go-deploy/service/v1/vms/cs_service"
	"go-deploy/service/v1/vms/opts"
	"sort"
)

// GetSnapshot gets a snapshot
func (c *Client) GetSnapshot(vmID string, id string, opts ...opts.GetSnapshotOpts) (*model.Snapshot, error) {
	_ = utils.GetFirstOrDefault(opts)

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

	return &model.Snapshot{
		ID:         snapshot.ID,
		VmID:       vmID,
		Name:       snapshot.Name,
		ParentName: snapshot.ParentName,
		CreatedAt:  snapshot.CreatedAt,
		State:      snapshot.State,
		Current:    snapshot.Current,
	}, nil
}

// GetSnapshotByName gets a snapshot by name
func (c *Client) GetSnapshotByName(vmID string, name string, opts ...opts.GetSnapshotOpts) (*model.Snapshot, error) {
	_ = utils.GetFirstOrDefault(opts)

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

	return &model.Snapshot{
		ID:         snapshot.ID,
		VmID:       vmID,
		Name:       snapshot.Name,
		ParentName: snapshot.ParentName,
		CreatedAt:  snapshot.CreatedAt,
		State:      snapshot.State,
		Current:    snapshot.Current,
	}, nil
}

// ListSnapshots lists snapshots
func (c *Client) ListSnapshots(vmID string, opts ...opts.ListSnapshotOpts) ([]model.Snapshot, error) {
	_ = utils.GetFirstOrDefault(opts)

	vm, err := c.Get(vmID)
	if err != nil {
		return nil, err
	}

	if vm == nil {
		return nil, nil
	}

	snapshots := make([]model.Snapshot, 0)
	for _, snapshot := range vm.Subsystems.CS.SnapshotMap {
		snapshots = append(snapshots, model.Snapshot{
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

// CreateSnapshot creates a snapshot
func (c *Client) CreateSnapshot(vmID string, opts *opts.CreateSnapshotOpts) error {
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
		return makeError(fmt.Errorf("vm not ready"))
	}

	params := &model.CreateSnapshotParams{}
	if opts.System != nil {
		params = opts.System
	} else if opts.User != nil {
		params.FromDTOv1(opts.User)
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

// DeleteSnapshot deletes a snapshot
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

// ApplySnapshot applies a snapshot
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

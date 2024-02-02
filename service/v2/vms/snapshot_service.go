package vms

import (
	"fmt"
	vmModels "go-deploy/models/sys/vm"
	"go-deploy/service/utils"
	"go-deploy/service/v2/vms/opts"
	"log"
	"sort"
)

// GetSnapshot gets a snapshot
func (c *Client) GetSnapshot(vmID string, id string, opts ...opts.GetSnapshotOpts) (*vmModels.SnapshotV2, error) {
	_ = utils.GetFirstOrDefault(opts)

	vm, err := c.Get(vmID)
	if err != nil {
		return nil, err
	}

	if vm == nil {
		return nil, nil
	}

	snapshot := vm.Subsystems.K8s.GetVmSnapshotByID(id)
	if snapshot == nil {
		return nil, nil
	}

	return &vmModels.SnapshotV2{
		ID:        snapshot.ID,
		Name:      snapshot.Name,
		Status:    snapshot.Status,
		CreatedAt: snapshot.CreatedAt,
	}, nil
}

// GetSnapshotByName gets a snapshot by name
func (c *Client) GetSnapshotByName(vmID string, name string, opts ...opts.GetSnapshotOpts) (*vmModels.SnapshotV2, error) {
	_ = utils.GetFirstOrDefault(opts)

	vm, err := c.Get(vmID)
	if err != nil {
		return nil, err
	}

	if vm == nil {
		return nil, nil
	}

	snapshot := vm.Subsystems.K8s.GetVmSnapshotByName(name)
	if snapshot == nil {
		return nil, nil
	}

	return &vmModels.SnapshotV2{
		ID:        snapshot.ID,
		Name:      snapshot.Name,
		Status:    snapshot.Status,
		CreatedAt: snapshot.CreatedAt,
	}, nil
}

// ListSnapshots lists snapshots
func (c *Client) ListSnapshots(vmID string, opts ...opts.ListSnapshotOpts) ([]vmModels.SnapshotV2, error) {
	_ = utils.GetFirstOrDefault(opts)

	vm, err := c.Get(vmID)
	if err != nil {
		return nil, err
	}

	if vm == nil {
		return nil, nil
	}

	snapshots := make([]vmModels.SnapshotV2, 0)
	for _, snapshot := range vm.Subsystems.K8s.GetVmSnapshotMap() {
		snapshots = append(snapshots, vmModels.SnapshotV2{
			ID:        snapshot.ID,
			Name:      snapshot.Name,
			Status:    snapshot.Status,
			CreatedAt: snapshot.CreatedAt,
		})
	}

	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].CreatedAt.Before(snapshots[j].CreatedAt)
	})

	return snapshots, nil
}

// CreateSnapshot creates a snapshot
func (c *Client) CreateSnapshot(vmID string, opts *opts.CreateSnapshotOpts) (*vmModels.SnapshotV2, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create snapshot for vm %s. details: %w", vmID, err)
	}

	vm, err := c.Get(vmID)
	if err != nil {
		return nil, makeError(err)
	}

	if vm == nil {
		log.Println("vm", vmID, "not found when creating snapshot. assuming it was deleted")
		return nil, nil
	}

	if !vm.Ready() {
		return nil, fmt.Errorf("vm %s not ready", vmID)
	}

	params := &vmModels.CreateSnapshotParams{}
	if opts.System != nil {
		params = opts.System
	} else if opts.User != nil {
		params.FromDTOv2(opts.User)
	}

	if params.Name == "" {
		log.Println("no snapshot type specified when creating snapshot for vm", vmID, ". did you forget to specify the type?")
		return nil, nil
	}

	snapshot, err := c.K8s().CreateVmSnapshot(vmID, params)
	if err != nil {
		return nil, makeError(err)
	}

	return snapshot, nil
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

	err = c.K8s().DeleteVmSnapshot(vm.ID, snapshotID)
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

	//log.Println("applying snapshot", snapshotID, "to vm", id)
	//err := c.K8s().ApplyVmSnapshot(id, snapshotID)
	//if err != nil {
	//	return makeError(err)
	//}

	return makeError(fmt.Errorf("not implemented"))
}

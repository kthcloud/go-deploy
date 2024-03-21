package snapshots

import (
	"fmt"
	"go-deploy/models/model"
	"go-deploy/service/utils"
	"go-deploy/service/v2/vms/opts"
	"log"
	"sort"
)

// Get gets a snapshot
func (c *Client) Get(vmID string, id string, opts ...opts.GetSnapshotOpts) (*model.SnapshotV2, error) {
	_ = utils.GetFirstOrDefault(opts)

	vm, err := c.VM(vmID, nil)
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

	return &model.SnapshotV2{
		ID:        snapshot.ID,
		Name:      snapshot.Name,
		Status:    snapshot.Status,
		CreatedAt: snapshot.CreatedAt,
	}, nil
}

// GetByName gets a snapshot by name
func (c *Client) GetByName(vmID string, name string, opts ...opts.GetSnapshotOpts) (*model.SnapshotV2, error) {
	_ = utils.GetFirstOrDefault(opts)

	vm, err := c.VM(vmID, nil)
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

	return &model.SnapshotV2{
		ID:        snapshot.ID,
		Name:      snapshot.Name,
		Status:    snapshot.Status,
		CreatedAt: snapshot.CreatedAt,
	}, nil
}

// List lists snapshots
func (c *Client) List(vmID string, opts ...opts.ListSnapshotOpts) ([]model.SnapshotV2, error) {
	_ = utils.GetFirstOrDefault(opts)

	vm, err := c.VM(vmID, nil)
	if err != nil {
		return nil, err
	}

	if vm == nil {
		return nil, nil
	}

	snapshots := make([]model.SnapshotV2, 0)
	for _, snapshot := range vm.Subsystems.K8s.GetVmSnapshotMap() {
		snapshots = append(snapshots, model.SnapshotV2{
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

// Create creates a snapshot
func (c *Client) Create(vmID string, opts ...opts.CreateSnapshotOpts) (*model.SnapshotV2, error) {
	o := utils.GetFirstOrDefault(opts)

	makeError := func(err error) error {
		return fmt.Errorf("failed to create snapshot for vm %s. details: %w", vmID, err)
	}

	vm, err := c.VM(vmID, nil)
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

	params := &model.CreateSnapshotParams{}
	if o.System != nil {
		params = o.System
	} else if o.User != nil {
		params.FromDTOv2(o.User)
	}

	if params.Name == "" {
		log.Println("no snapshot type specified when creating snapshot for vm", vmID, ". did you forget to specify the type?")
		return nil, nil
	}

	snapshot, err := c.V2.VMs().K8s().CreateVmSnapshot(vmID, params)
	if err != nil {
		return nil, makeError(err)
	}

	return snapshot, nil
}

// Delete deletes a snapshot
func (c *Client) Delete(vmID, snapshotID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete snapshot %s from vm %s. details: %w", snapshotID, vmID, err)
	}

	vm, err := c.VM(vmID, nil)
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

	err = c.V2.VMs().K8s().DeleteVmSnapshot(vm.ID, snapshotID)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// Apply applies a snapshot
func (c *Client) Apply(vmID, snapshotID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to apply snapshot %s to vm %s. details: %w", snapshotID, vmID, err)
	}

	//log.Println("applying snapshot", snapshotID, "to vm", id)
	//err := c.K8s().ApplyVmSnapshot(id, snapshotID)
	//if err != nil {
	//	return makeError(err)
	//}

	return makeError(fmt.Errorf("not implemented"))
}

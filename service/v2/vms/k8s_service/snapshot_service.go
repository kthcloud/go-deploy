package k8s_service

import (
	"fmt"
	"go-deploy/models/model"
	"go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service/resources"
)

func (c *Client) CreateVmSnapshot(vmID string, params *model.CreateSnapshotParams) (*model.SnapshotV2, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create snapshot for k8s vm %s. details: %w", vmID, err)
	}

	vm, err := c.VM(vmID, nil)
	if err != nil {
		return nil, err
	}

	if vm == nil {
		return nil, nil
	}

	vm, kc, _, err := c.Get(OptsNoGenerator(vmID))
	if err != nil {
		return nil, err
	}

	var k8sVmID string
	if k8sVM := vm.Subsystems.K8s.GetVM(vm.Name); k8sVM != nil {
		k8sVmID = k8sVM.ID
	} else {
		return nil, nil
	}

	snapshotPublic := &models.VmSnapshotPublic{
		Name:      params.Name,
		Namespace: kc.Namespace,
		VmID:      k8sVmID,
	}

	err = resources.SsCreator(kc.CreateVmSnapshot).
		WithDbFunc(dbFunc(vmID, "vmSnapshotMap."+params.Name)).
		WithPublic(snapshotPublic).
		Exec()
	if err != nil {
		return nil, makeError(err)
	}

	return &model.SnapshotV2{
		ID:        snapshotPublic.ID,
		Name:      snapshotPublic.Name,
		CreatedAt: snapshotPublic.CreatedAt,
		Status:    "created",
	}, nil
}

func (c *Client) DeleteVmSnapshot(vmID string, id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete snapshot for k8s vm %s. details: %w", vmID, err)
	}

	vm, kc, _, err := c.Get(OptsNoGenerator(vmID))
	if err != nil {
		return err
	}

	if vm == nil {
		return nil
	}

	for mapName, snapshot := range vm.Subsystems.K8s.GetVmSnapshotMap() {
		if snapshot.ID == id {
			err = resources.SsDeleter(kc.DeleteVmSnapshot).
				WithDbFunc(dbFunc(vmID, "vmSnapshotMap."+mapName)).
				WithResourceID(id).
				Exec()
			if err != nil {
				return makeError(err)
			}
			break
		}
	}

	return nil
}

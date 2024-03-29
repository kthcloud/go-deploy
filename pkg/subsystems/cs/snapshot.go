package cs

import (
	"fmt"
	"go-deploy/pkg/subsystems/cs/models"
	"log"
	"strings"
)

// ReadSnapshot reads the snapshot from CloudStack.
func (client *Client) ReadSnapshot(id string) (*models.SnapshotPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read snapshot %s. details: %w", id, err)
	}

	if id == "" {
		log.Println("cs snapshot not supplied when reading. assuming it was deleted")
		return nil, nil
	}

	params := client.CsClient.Snapshot.NewListVMSnapshotParams()
	params.SetProjectid(client.ProjectID)
	params.SetListall(true)

	listResponse, err := client.CsClient.Snapshot.ListVMSnapshot(params)
	if err != nil {
		return nil, makeError(err)
	}

	var snapshot *models.SnapshotPublic
	for _, s := range listResponse.VMSnapshot {
		if s.Id == id {
			snapshot = models.CreateSnapshotPublicFromGet(s)
		}
	}

	return snapshot, nil
}

// ReadAllSnapshots reads all snapshots from CloudStack.
func (client *Client) ReadAllSnapshots(vmID string) ([]models.SnapshotPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read all snapshots. details: %w", err)
	}

	params := client.CsClient.Snapshot.NewListVMSnapshotParams()
	params.SetProjectid(client.ProjectID)
	params.SetVirtualmachineid(vmID)
	params.SetListall(true)

	listResponse, err := client.CsClient.Snapshot.ListVMSnapshot(params)
	if err != nil {
		return nil, makeError(err)
	}

	var snapshots []models.SnapshotPublic
	for _, s := range listResponse.VMSnapshot {
		snapshots = append(snapshots, *models.CreateSnapshotPublicFromGet(s))
	}

	return snapshots, nil
}

// CreateSnapshot creates the snapshot in CloudStack.
func (client *Client) CreateSnapshot(public *models.SnapshotPublic) (*models.SnapshotPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create snapshot %s. details: %w", public.Name, err)
	}

	params := client.CsClient.Snapshot.NewCreateVMSnapshotParams(public.VmID)
	params.SetSnapshotmemory(true) // required by KVM
	params.SetVirtualmachineid(public.VmID)
	params.SetName(public.Name)
	params.SetDescription(public.Description)
	params.SetQuiescevm(false)

	createResponse, err := client.CsClient.Snapshot.CreateVMSnapshot(params)
	if err != nil {
		if strings.Contains(err.Error(), "There is other active vm snapshot tasks on the instance") {
			log.Println("other snapshots are being created for vm", public.VmID, ". must wait for them to finish first")
			return nil, nil
		}

		if strings.Contains(err.Error(), "Domain not found") {
			log.Println("cs vm not found. assuming it was deleted")
			return nil, nil
		}

		return nil, makeError(err)
	}

	return models.CreateSnapshotPublicFromCreate(createResponse), nil
}

// DeleteSnapshot deletes the snapshot in CloudStack.
func (client *Client) DeleteSnapshot(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete snapshot %s. details: %w", id, err)
	}

	if id == "" {
		log.Println("cs snapshot not supplied when deleting. assuming it was deleted")
		return nil
	}

	params := client.CsClient.Snapshot.NewDeleteVMSnapshotParams(id)

	_, err := client.CsClient.Snapshot.DeleteVMSnapshot(params)
	if err != nil {
		if !strings.Contains(err.Error(), "entity does not exist") && !strings.Contains(err.Error(), "Unable to find") {
			return makeError(err)
		}
	}

	return nil
}

// ApplySnapshot applies the snapshot in CloudStack.
func (client *Client) ApplySnapshot(public *models.SnapshotPublic) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to apply snapshot %s. details: %w", public.Name, err)
	}

	params := client.CsClient.Snapshot.NewRevertToVMSnapshotParams(public.ID)

	_, err := client.CsClient.Snapshot.RevertToVMSnapshot(params)
	if err != nil {
		return makeError(err)
	}

	return nil
}

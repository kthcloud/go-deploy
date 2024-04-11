package gpu_repo

import (
	"context"
	"errors"
	"fmt"
	"go-deploy/models/model"
	"go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

// Create creates a new gpu_repo in the database.
// If the GPU already exists, it does nothing.
func (client *Client) Create(id, host string, data model.GpuData, zone string) error {
	currentGPU, err := client.GetByID(id)
	if err != nil {
		return err
	}

	if currentGPU != nil {
		return nil
	}

	gpu := model.GPU{
		ID:   id,
		Host: host,
		Lease: model.Lease{
			VmID:   "",
			UserID: "",
			End:    time.Time{},
		},
		Data: data,
		Zone: zone,
	}

	_, err = client.Collection.InsertOne(context.TODO(), gpu)
	if err != nil {
		// If the error is a duplicate key error, it means that another instance already created the GPU
		// Probably caused by a race condition
		if mongo.IsDuplicateKeyError(err) {
			return nil
		}

		err = fmt.Errorf("failed to create gpu. details: %w", err)
		return err
	}

	return nil
}

func (client *Client) Attach(gpuID, vmID, user string, end time.Time) (bool, error) {
	gpu, err := client.GetByID(gpuID)
	if err != nil {
		return false, err
	}

	if gpu == nil {
		return false, NotFoundErr
	}

	if gpu.Lease.VmID != "" && gpu.Lease.VmID != vmID {
		return false, AlreadyAttachedErr
	}

	// First check if the GPU is already attached to this vm
	if gpu.Lease.VmID == vmID {
		if gpu.Lease.IsExpired() {
			// renew lease
			filter := bson.D{{"id", gpuID}, {"lease.vmId", vmID}}
			update := bson.D{{"lease.end", end}}

			err = client.SetWithBsonByFilter(filter, update)
			if err != nil {
				if errors.Is(err, mongo.ErrNoDocuments) {
					// This is not treated as an error, just another instance snatched the GPU before this one
					return false, nil
				}

				err = fmt.Errorf("failed to update gpu. details: %w", err)
				return false, err
			}
		}

		// Either way return true, since a renewal succeeded or nothing happened (still attached)
		return true, nil
	}

	// If this is not a renewal, try to attach the GPU to the vm
	if !gpu.IsAttached() {
		filter := bson.D{
			{"id", gpuID},
			{"$or", []interface{}{
				bson.M{"lease.vmId": ""},
				bson.M{"lease": bson.M{"$exists": false}},
			}}}
		update := bson.D{
			{"lease.vmId", vmID},
			{"lease.user", user},
			{"lease.end", end},
		}

		err = client.SetWithBsonByFilter(filter, update)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				// This is not treated as an error, just another instance snatched the GPU before this one
				return false, nil
			}
			return false, err
		}
	}

	return true, nil
}

func (client *Client) Detach(vmID string) error {
	gpu, err := client.WithVM(vmID).Get()
	if err != nil {
		return err
	}

	if gpu == nil {
		// already detached
		return nil
	}

	filter := bson.D{
		{"id", gpu.ID},
		{"lease.vmId", vmID},
	}

	update := bson.D{
		{"lease.vmId", ""},
		{"lease.user", ""},
		{"lease.end", time.Time{}},
	}

	err = client.SetWithBsonByFilter(filter, update)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return err
		}

		utils.PrettyPrintError(fmt.Errorf("failed to clear gpu lease for vm %s. details: %w", vmID, err))
	}

	return nil
}

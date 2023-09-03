package gpu

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"go-deploy/models/dto/body"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func (gpu *GPU) ToDTO(addUserInfo bool) body.GpuRead {
	id := base64.StdEncoding.EncodeToString([]byte(gpu.ID))

	var lease *body.GpuLease

	if gpu.Lease.VmID != "" {
		lease = &body.GpuLease{
			End:     gpu.Lease.End,
			Expired: gpu.Lease.IsExpired(),
		}

		if addUserInfo {
			lease.User = &gpu.Lease.UserID
			lease.VmID = &gpu.Lease.VmID
		}
	}

	return body.GpuRead{
		ID:    id,
		Name:  gpu.Data.Name,
		Lease: lease,
	}
}

func (client *Client) Create(id, host string, data GpuData, zone string) error {
	currentGPU, err := client.GetByID(id)
	if err != nil {
		return err
	}

	if currentGPU != nil {
		return nil
	}

	gpu := GPU{
		ID:   id,
		Host: host,
		Lease: GpuLease{
			VmID:   "",
			UserID: "",
			End:    time.Time{},
		},
		Data: data,
		Zone: zone,
	}

	_, err = client.Collection.InsertOne(context.TODO(), gpu)
	if err != nil {
		err = fmt.Errorf("failed to create gpu. details: %w", err)
		return err
	}

	return nil
}

func (client *Client) GetByID(id string) (*GPU, error) {
	var gpu GPU
	err := client.Collection.FindOne(context.TODO(), bson.D{{"id", id}}).Decode(&gpu)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		err = fmt.Errorf("failed to fetch gpu. details: %w", err)
		return nil, err
	}

	return &gpu, err
}

func (client *Client) GetAll() ([]GPU, error) {
	filter := bson.D{
		{"host", bson.M{"$nin": client.ExcludedHosts}},
		{"id", bson.M{"$nin": client.ExcludedGPUs}},
	}

	var gpus []GPU
	cursor, err := client.Collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	err = cursor.All(context.Background(), &gpus)
	if err != nil {
		return nil, err
	}

	return gpus, nil
}

func (client *Client) GetAllLeased() ([]GPU, error) {
	// filter lease exist and vmId is not empty
	filter := bson.D{
		{"$and", []interface{}{
			bson.M{"lease.vmId": bson.M{"$ne": ""}},
			bson.M{"lease": bson.M{"$exists": true}},
		}},
		{"host", bson.M{"$nin": client.ExcludedHosts}},
		{"id", bson.M{"$nin": client.ExcludedGPUs}},
	}

	var gpus []GPU
	cursor, err := client.Collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	err = cursor.All(context.Background(), &gpus)
	if err != nil {
		return nil, err
	}

	return gpus, nil
}

func (client *Client) GetAllAvailable() ([]GPU, error) {
	now := time.Now()

	filter := bson.D{
		{"$or", []interface{}{
			bson.M{"lease": bson.M{"$exists": false}},
			bson.M{"lease.vmId": ""},
			bson.M{"lease.end": bson.M{"$lte": now}},
		}},
		{"host", bson.M{"$nin": client.ExcludedHosts}},
		{"id", bson.M{"$nin": client.ExcludedGPUs}},
	}

	var gpus []GPU
	cursor, err := client.Collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	err = cursor.All(context.Background(), &gpus)
	if err != nil {
		return nil, err
	}

	return gpus, nil
}

func (client *Client) Delete(gpuID string) error {
	err := client.Collection.FindOneAndDelete(context.Background(), bson.D{{"id", gpuID}}).Err()
	if err != nil {
		return fmt.Errorf("failed to delete gpu. details: %w", err)
	}

	return nil
}

func (client *Client) Attach(gpuID, vmID, user string, end time.Time) (bool, error) {
	vm, err := vmModel.New().GetByID(vmID)
	if err != nil {
		return false, err
	}

	if vm == nil {
		return false, nil
	}

	gpu, err := client.GetByID(gpuID)
	if err != nil {
		return false, err
	}

	if gpu == nil {
		return false, fmt.Errorf("gpu not found")
	}

	if gpu.Lease.VmID != "" && gpu.Lease.VmID != vmID {
		return false, fmt.Errorf("gpu is already attached to another vm")
	}

	// first check if the gpu is already attached to this vm
	if gpu.Lease.VmID == vmID {
		if gpu.Lease.IsExpired() {
			// renew lease
			filter := bson.D{
				{"id", gpuID},
				{"lease.vmId", vmID},
			}
			update := bson.M{
				"$set": bson.M{
					"lease.end": end,
				},
			}

			opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

			err = client.Collection.FindOneAndUpdate(context.Background(), filter, update, opts).Decode(&gpu)
			if err != nil {
				if errors.Is(err, mongo.ErrNoDocuments) {
					// this is not treated as an error, just another instance snatched the gpu before this one
					return false, nil
				}

				err = fmt.Errorf("failed to update gpu. details: %w", err)
				return false, err
			}
		}

		// either way return true, since a renewal succeeded or nothing happened (still attached)
		return true, nil
	}

	// if this is not a renewal, try to attach the gpu to the vm
	if !gpu.IsAttached() {
		filter := bson.D{
			{"id", gpuID},
			{"$or", []interface{}{
				bson.M{"lease.vmId": ""},
				bson.M{"lease": bson.M{"$exists": false}},
			}}}
		update := bson.M{
			"$set": bson.M{
				"lease.vmId": vmID,
				"lease.user": user,
				"lease.end":  end,
			},
		}

		opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

		err = client.Collection.FindOneAndUpdate(context.Background(), filter, update, opts).Decode(&gpu)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				// this is not treated as an error, just another instance snatched the gpu before this one
				return false, nil
			}
			return false, err
		}
	}

	err = vmModel.New().UpdateWithBsonByID(vmID, bson.D{{"gpuId", gpuID}})
	if err != nil {
		// remove lease, if this also fails, we are in a bad state...
		_, _ = client.Collection.UpdateOne(
			context.TODO(),
			bson.D{{"id", gpuID}},
			bson.M{"$set": bson.M{"lease": GpuLease{}}},
		)
		err := fmt.Errorf("failed to remove lease after vm update failed. system is now in an inconsistent state. please fix manually. vm id: %s gpu id: %s. details: %w", vmID, gpuID, err)
		utils.PrettyPrintError(err)
		return false, err
	}

	return true, nil
}

func (client *Client) Detach(vmID, userID string) error {
	vm, err := vmModel.New().GetByID(vmID)
	if err != nil {
		return err
	}

	if vm == nil {
		return nil
	}

	if vm.GpuID == "" {
		return nil
	}

	gpu, err := client.GetByID(vm.GpuID)
	if err != nil {
		return err
	}

	if gpu == nil {
		return fmt.Errorf("gpu not found")
	}

	if gpu.Lease.VmID != vmID {
		return fmt.Errorf("vm is not attached to this gpu")
	}

	filter := bson.D{
		{"id", gpu.ID},
		{"lease.vmId", vmID},
		{"lease.user", userID},
	}

	update := createClearedLeaseFilter()

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	err = client.Collection.FindOneAndUpdate(context.Background(), filter, update, opts).Decode(&gpu)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// this is not treated as an error, just another instance snatched the gpu before this one
			return nil
		}
		return err
	}

	err = vmModel.New().UpdateWithBsonByID(vmID, bson.D{{"gpuId", ""}})
	if err != nil {
		// remove lease, if this also fails, we are in a bad state...
		_, _ = client.Collection.UpdateOne(
			context.TODO(),
			bson.D{{"id", gpu.ID}},
			bson.M{"$set": bson.M{"lease": GpuLease{}}},
		)
		err := fmt.Errorf("failed to remove lease after vm update failed. system is now in an inconsistent state. please fix manually. vm id: %s gpu id: %s. details: %w", vmID, gpu.ID, err)
		utils.PrettyPrintError(err)
	}

	return nil
}

func (client *Client) ClearLease(gpuID string) error {
	filter := bson.D{
		{"id", gpuID},
	}

	update := createClearedLeaseFilter()

	err := client.Collection.FindOneAndUpdate(context.Background(), filter, update, nil).Err()
	if err != nil {
		return err
	}

	return nil
}

func createClearedLeaseFilter() bson.M {
	return bson.M{
		"$set": bson.M{
			"lease.vmId": "",
			"lease.user": "",
			"lease.end":  time.Time{},
		},
	}
}

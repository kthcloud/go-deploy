package vm

import (
	"context"
	"encoding/base64"
	"fmt"
	"go-deploy/models"
	"go-deploy/models/dto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type GpuData struct {
	Name     string `bson:"name" json:"name"`
	Slot     string `bson:"slot" json:"slot"`
	Vendor   string `bson:"vendor" json:"vendor"`
	VendorID string `bson:"vendorId" json:"vendorId"`
	Bus      string `bson:"bus" json:"bus"`
	DeviceID string `bson:"deviceId" json:"deviceId"`
}

type GpuLease struct {
	VmID string    `bson:"vmId" json:"vmId"`
	User string    `bson:"user" json:"user"`
	End  time.Time `bson:"end" json:"end"`
}

type GPU struct {
	ID    string   `bson:"id" json:"id"`
	Host  string   `bson:"host" json:"host"`
	Lease GpuLease `bson:"lease" json:"lease"`
	Data  GpuData  `bson:"data" json:"data"`
}

func (gpu *GPU) ToDto() dto.GpuRead {
	id := base64.StdEncoding.EncodeToString([]byte(gpu.ID))

	var lease *dto.GpuLease

	if gpu.Lease.VmID != "" {
		lease = &dto.GpuLease{
			VmID: gpu.Lease.VmID,
			User: gpu.Lease.User,
			End:  gpu.Lease.End,
		}
	}

	return dto.GpuRead{
		ID:    id,
		Name:  gpu.Data.Name,
		Lease: lease,
	}
}

func CreateGPU(id, host string, data GpuData) error {
	currentGPU, err := GetGpuByID(id)
	if err != nil {
		return err
	}

	if currentGPU != nil {
		return nil
	}

	vm := GPU{
		ID:   id,
		Host: host,
		Lease: GpuLease{
			VmID: "",
			User: "",
			End:  time.Time{},
		},
		Data: data,
	}

	_, err = models.GpuCollection.InsertOne(context.TODO(), vm)
	if err != nil {
		err = fmt.Errorf("failed to create gpu. details: %s", err)
		return err
	}

	return nil
}

func GetGpuByID(id string) (*GPU, error) {
	var gpu GPU
	err := models.GpuCollection.FindOne(context.TODO(), bson.D{{"id", id}}).Decode(&gpu)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}

		err = fmt.Errorf("failed to fetch vm. details: %s", err)
		return nil, err
	}

	return &gpu, err
}

func GetAllGPUs() ([]GPU, error) {
	var gpus []GPU
	cursor, err := models.GpuCollection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}

	err = cursor.All(context.Background(), &gpus)
	if err != nil {
		return nil, err
	}

	return gpus, nil
}

func GetAllAvailableGPUs() ([]GPU, error) {
	var gpus []GPU
	cursor, err := models.GpuCollection.Find(context.Background(), bson.D{{"lease.vmId", ""}})
	if err != nil {
		return nil, err
	}

	err = cursor.All(context.Background(), &gpus)
	if err != nil {
		return nil, err
	}

	return gpus, nil
}

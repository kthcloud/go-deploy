package vm

import (
	"context"
	"fmt"
	"go-deploy/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type GpuData struct {
	Name     string `bson:"name" json:"name"`
	Slot     string `bson:"slot" json:"slot"`
	Vendor   string `bson:"vendor" json:"vendor"`
	VendorID string `bson:"vendorId" json:"vendorId"`
	Bus      string `bson:"bus" json:"bus"`
	DeviceID string `bson:"deviceId" json:"deviceId"`
}

type GPU struct {
	ID   string  `bson:"id" json:"id"`
	Host string  `bson:"host" json:"host"`
	VM   string  `bson:"vm" json:"vm"`
	Data GpuData `bson:"data" json:"data"`
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
		VM:   "",
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

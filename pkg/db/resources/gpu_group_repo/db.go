package gpu_group_repo

import (
	"errors"
	"go-deploy/models/model"
	"go-deploy/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
)

func (client *Client) Create(id, zone, vendor, deviceID, vendorID string, total int) error {
	group := model.GpuGroup{
		ID:       id,
		Zone:     zone,
		Total:    total,
		Vendor:   vendor,
		DeviceID: deviceID,
		VendorID: vendorID,
	}

	// We assume there is a unique constraint on name + zone
	err := client.CreateIfUnique(id, &group, bson.D{{"id", id}, {"zone", zone}})
	if err != nil {
		if errors.Is(err, db.UniqueConstraintErr) {
			return GpuLeaseAlreadyExistsErr
		}

		return err
	}

	return nil
}

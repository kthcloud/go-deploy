package gpu_group_repo

import (
	"errors"
	"fmt"
	"go-deploy/models/model"
	"go-deploy/pkg/db"
	"go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
)

func (client *Client) Create(name, displayName, zone, vendor, deviceID, vendorID string, total int) error {
	id := utils.HashStringAlphanumericLower(fmt.Sprintf("%s-%s", name, zone))

	group := model.GpuGroup{
		ID:          id,
		Name:        name,
		DisplayName: displayName,
		Zone:        zone,
		Total:       total,
		Vendor:      vendor,
		DeviceID:    deviceID,
		VendorID:    vendorID,
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

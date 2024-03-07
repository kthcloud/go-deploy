package gpu_lease

import (
	"errors"
	"go-deploy/models"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func (client *Client) Create(id, vmID, userID, groupName string, leaseDuration float64) error {
	lease := GpuLease{
		ID:            id,
		GroupName:     groupName,
		VmID:          vmID,
		UserID:        userID,
		LeaseDuration: leaseDuration,
		ActivatedAt:   nil,
		CreatedAt:     time.Now(),
	}

	// Right now we only allow one lease per user. We assume there is a unique index set up for this.
	err := client.CreateIfUnique(id, &lease, bson.D{{"id", id}, {"userId", userID}})
	if err != nil {
		if errors.Is(err, models.UniqueConstraintErr) {
			return GpuLeaseAlreadyExistsErr
		}

		return err
	}

	return nil
}

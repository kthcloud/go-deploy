package gpu_lease_repo

import (
	"errors"
	"time"

	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
)

func (client *Client) Create(id, userID, groupName string, leaseDuration float64) error {

	lease := model.GpuLease{
		ID:            id,
		GpuGroupID:    groupName,
		VmID:          nil,
		UserID:        userID,
		LeaseDuration: leaseDuration,
		ActivatedAt:   nil,
		AssignedAt:    nil,
		ExpiredAt:     nil,
		CreatedAt:     time.Now(),
	}

	// Right now we only allow one lease per user. We assume there is a unique index set up for this.
	err := client.CreateIfUnique(id, &lease, bson.D{{Key: "userId", Value: userID}})
	if err != nil {
		if errors.Is(err, db.ErrUniqueConstraint) {
			return ErrGpuLeaseAlreadyExists
		}

		return err
	}

	return nil
}

func (client *Client) UpdateWithParams(id string, params *model.GpuLeaseUpdateParams) error {
	update := bson.D{}

	db.AddIfNotNil(&update, "activatedAt", params.ActivatedAt)
	db.AddIfNotNil(&update, "vmId", params.VmID)

	err := client.SetWithBsonByID(id, update)
	if err != nil {
		if errors.Is(err, db.ErrUniqueConstraint) {
			return ErrVmAlreadyAttached
		}

		return err
	}

	return nil
}

func (client *Client) Release() error {
	return client.SetWithBSON(bson.D{{Key: "vmId", Value: nil}})
}

func (client *Client) ReleaseByID(id string) error {
	return client.SetWithBsonByID(id, bson.D{{Key: "vmId", Value: nil}})
}

func (client *Client) SetExpiry(id string, expiresAt time.Time) error {
	return client.SetWithBsonByID(id, bson.D{{Key: "expiresAt", Value: expiresAt}})
}

func (client *Client) MarkExpired(id string) error {
	return client.SetWithBsonByID(id, bson.D{{Key: "expiredAt", Value: time.Now()}})
}

func (client *Client) MarkAssigned(id string) error {
	return client.SetWithBsonByID(id, bson.D{{Key: "assignedAt", Value: time.Now()}})
}

func (client *Client) MarkActivated(id string) error {
	return client.SetWithBsonByID(id, bson.D{{Key: "activatedAt", Value: time.Now()}})
}

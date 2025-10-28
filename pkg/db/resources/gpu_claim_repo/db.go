package gpu_claim_repo

import (
	"errors"
	"fmt"

	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
)

func (client *Client) Create(id string, params *model.GpuClaimCreateParams) error {
	if params == nil {
		return fmt.Errorf("params is nil")
	}

	claim := model.GpuClaim{
		ID:   id,
		Name: params.Name,
		Zone: params.Zone,
	}

	// We assume there is a unique constraint on id + zone + name
	err := client.CreateIfUnique(id, &claim, bson.D{{Key: "id", Value: id}, {Key: "zone", Value: params.Zone}, {Key: "name", Value: params.Name}})
	if err != nil {
		if errors.Is(err, db.ErrUniqueConstraint) {
			return ErrGpuClaimAlreadyExists
		}

		return err
	}

	return nil
}

// DeleteSubsystem deletes a subsystem from a gpu claim.
// It prepends the key with `subsystems` and unsets it.
func (client *Client) DeleteSubsystem(id, key string) error {
	subsystemKey := fmt.Sprintf("subsystems.%s", key)
	return client.UpdateWithBsonByID(id, bson.D{{Key: "$unset", Value: bson.D{{Key: subsystemKey, Value: ""}}}})
}

// SetSubsystem sets a subsystem in a gpu claim.
// It prepends the key with `subsystems` and sets it.
func (client *Client) SetSubsystem(id, key string, update any) error {
	subsystemKey := fmt.Sprintf("subsystems.%s", key)
	return client.SetWithBsonByID(id, bson.D{{Key: subsystemKey, Value: update}})
}

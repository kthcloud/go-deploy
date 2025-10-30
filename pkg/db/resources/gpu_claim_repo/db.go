package gpu_claim_repo

import (
	"errors"
	"fmt"
	"time"

	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db"
	"github.com/kthcloud/go-deploy/utils/convutils"
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
		Requested: convutils.ToNameMap(params.Requested, func(r model.RequestedGpuCreate) string { return r.Name }, func(r model.RequestedGpuCreate) model.RequestedGpu {
			return r.RequestedGpu
		}),
		Activities: make(map[string]model.Activity),
		Subsystems: model.GpuClaimSubsystems{},
		CreatedAt:  time.Now(),
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

type ReconcileOption func(d *bson.D)

func WithSetStatus(status *model.GpuClaimStatus) ReconcileOption {
	return func(d *bson.D) {
		(*d) = append(*d, bson.E{Key: "status", Value: status})
	}
}

func WithSetAllocated(allocated map[string]model.AllocatedGpu) ReconcileOption {
	return func(d *bson.D) {
		(*d) = append(*d, bson.E{Key: "allocated", Value: allocated})
	}
}

func WithSetConsumers(consumers []model.GpuClaimConsumer) ReconcileOption {
	return func(d *bson.D) {
		(*d) = append(*d, bson.E{Key: "consumers", Value: consumers})
	}
}

// ReconcileStateByName batches Reconcile options to update the state of a gpu claim.
func (client *Client) ReconcileStateByName(name string, opts ...ReconcileOption) error {
	update := bson.D{}

	for _, opt := range opts {
		opt(&update)
	}
	if len(update) < 1 {
		// no-op
		return nil
	}

	return client.SetWithBsonByName(name, update)
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

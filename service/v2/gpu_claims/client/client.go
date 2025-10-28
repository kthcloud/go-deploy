package client

import (
	"fmt"

	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db/resources/gpu_claim_repo"
	"github.com/kthcloud/go-deploy/service/core"
)

// BaseClient is the base client for all the subsystems client for GPUClaims.
type BaseClient[parent any] struct {
	p *parent

	// Cache is used to cache the resources fetched inside the service.
	Cache *core.Cache
}

// NewBaseClient creates a new BaseClient.
func NewBaseClient[parent any](cache *core.Cache) BaseClient[parent] {
	if cache == nil {
		cache = core.NewCache()
	}

	return BaseClient[parent]{Cache: cache}
}

// SetParent sets the parent of the client.
// This ensures the correct parent client is returned when calling builder methods.
func (c *BaseClient[parent]) SetParent(p *parent) {
	c.p = p
}

// GpuClaim returns the GpuClaim with the given ID.
// After a successful fetch, the GpuClaim will be cached.
func (c *BaseClient[parent]) GpuClaim(id string, gcc *gpu_claim_repo.Client) (*model.GpuClaim, error) {
	gc := c.Cache.GetGpuClaim(id)
	if gc != nil {
		return gc, nil
	}

	return c.fetchGpuClaim(id, gcc)
}

// GpuClaim returns the GpuClaim with the given ID.
// After a successful fetch, the GpuClaim will be cached.
func (c *BaseClient[parent]) GpuClaims(gcc *gpu_claim_repo.Client) ([]model.GpuClaim, error) {
	return c.fetchGpuClaims(gcc)
}

// Refresh clears the cache for the GpuClaim with the given ID and fetches it again.
// After a successful fetch, the GpuClaim is cached.
func (c *BaseClient[parent]) Refresh(id string) (*model.GpuClaim, error) {
	return c.fetchGpuClaim(id, nil)
}

func (c *BaseClient[parent]) fetchGpuClaims(gcc *gpu_claim_repo.Client) ([]model.GpuClaim, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch gpu claims in service client: %w", err)
	}

	if gcc == nil {
		gcc = gpu_claim_repo.New()
	}

	gcs, err := gcc.List()
	if err != nil {
		return nil, makeError(err)
	}

	for _, gc := range gcs {
		v := gc
		c.Cache.StoreGpuClaim(&v)
	}

	return gcs, nil
}

func (c *BaseClient[parent]) fetchGpuClaim(id string, gcc *gpu_claim_repo.Client) (*model.GpuClaim, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch gpu claim in service client: %w", err)
	}

	if gcc == nil {
		gcc = gpu_claim_repo.New()
	}

	var gc *model.GpuClaim
	var err error

	if id == "" {
		gc, err = gcc.Get()
	} else {
		gc, err = gcc.GetByID(id)
	}

	if err != nil {
		return nil, makeError(err)
	}

	if gc == nil {
		return nil, nil
	}

	c.Cache.StoreGpuClaim(gc)
	return gc, nil
}

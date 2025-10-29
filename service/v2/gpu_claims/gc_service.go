package gpu_claims

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/db/resources/gpu_claim_repo"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
	serviceUtils "github.com/kthcloud/go-deploy/service/utils"
	"github.com/kthcloud/go-deploy/service/v2/gpu_claims/k8s_service"
	"github.com/kthcloud/go-deploy/service/v2/gpu_claims/opts"
)

// Get detailed gpu claims
//
// Only admin users are allowed to get the detailed gpu claim
func (c *Client) Get(id string, opts ...opts.Opts) (*model.GpuClaim, error) {
	o := serviceUtils.GetFirstOrDefault(opts)

	gcrc := gpu_claim_repo.New()

	if c.V2.Auth() == nil || !c.V2.Auth().User.IsAdmin {
		return nil, sErrors.ErrForbidden
	}

	if o.Zone != nil {
		gcrc.WithZone(o.Zone.Name)
	}

	return c.GpuClaim(id, gcrc)
}
func (c *Client) List(opts ...opts.ListOpts) ([]model.GpuClaim, error) {
	o := serviceUtils.GetFirstOrDefault(opts)

	gcrc := gpu_claim_repo.New()

	if o.Pagination != nil {
		gcrc.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	return c.GpuClaims(gcrc)

}

// Create creates a new gpu claim
func (c *Client) Create(id string, params *model.GpuClaimCreateParams) error {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to create gpu claim. details: %w", err)
	}

	if params.Zone = strings.TrimSpace(params.Zone); params.Zone == "" {
		params.Zone = config.Config.Deployment.DefaultZone
	}

	err := gpu_claim_repo.New().Create(id, params)
	if err != nil {
		if errors.Is(err, gpu_claim_repo.ErrGpuClaimAlreadyExists) {
			return sErrors.ErrSmAlreadyExists
		}

		return makeErr(err)
	}

	err = k8s_service.New(c.Cache).Create(id, params)
	if err != nil {
		return makeErr(err)
	}

	return nil
}

// Delete deletes an existing gpu claim
func (c *Client) Delete(id string) error {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to delete gpu claim. details: %w", err)
	}

	log.Println("Deleting gpu claim", id)

	err := k8s_service.New(c.Cache).Delete(id)
	if err != nil {
		return makeErr(err)
	}

	return nil
}

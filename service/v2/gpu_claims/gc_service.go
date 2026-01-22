package gpu_claims

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"

	modelConfig "github.com/kthcloud/go-deploy/models/config"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/db/resources/gpu_claim_repo"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/parsers/dra/nvidia"
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

	if o.Zone != nil {
		gcrc.WithZone(*o.Zone)
	}

	if o.Roles != nil && !slices.Contains(*o.Roles, "admin") {
		gcrc.WithRoles(*o.Roles)
	}

	if o.Names != nil {
		gcrc.WithNames(*o.Names)
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

	if !c.V2.System().ZoneHasCapability(params.Zone, modelConfig.ZoneCapabilityDRA) {
		return sErrors.NewZoneCapabilityMissingError(params.Zone, modelConfig.ZoneCapabilityDRA)
	}

	if len(params.Requested) < 1 {
		return sErrors.ErrBadGpuClaimNoRequest
	} else {
		for _, req := range params.Requested {
			if req.Config != nil {
				if strings.TrimSpace(req.Config.Driver) == "" {
					return makeErr(fmt.Errorf("config provided but driver is empty"))
				} else {
					log.Println("driver: ", req.Config.Driver)
				}
			}
		}
	}

	err := gpu_claim_repo.New().Create(id, params)
	if err != nil {
		if errors.Is(err, gpu_claim_repo.ErrGpuClaimAlreadyExists) {
			return sErrors.ErrGpuClaimAlreadyExists
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

	repo := gpu_claim_repo.New()

	claim, err := repo.GetByID(id)
	if err != nil || claim == nil {
		if err != nil {
			return makeErr(err)
		}
		if claim == nil {
			return makeErr(sErrors.ErrResourceNotFound)
		}
	}

	// TODO: cascade references, schedule update jobs that remove the gpuClaim requests.
	/*depls, err := c.V2.Deployments().List(deploymentOpts.ListOpts{
		GpuClaim: claim.Name,
	})*/

	err = k8s_service.New(c.Cache).Delete(id)
	if err != nil {
		return makeErr(err)
	}

	if err := repo.DeleteByID(id); err != nil {
		return makeErr(err)
	}

	return nil
}

func (c *Client) Update(id string, params *model.GpuClaimUpdateParams) error {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to update gpu claim. details: %w", err)
	}
	repo := gpu_claim_repo.New()

	claim, err := repo.GetByID(id)
	if err != nil || claim == nil {
		if err != nil {
			return makeErr(err)
		}
		if claim == nil {
			return makeErr(sErrors.ErrResourceNotFound)
		}
	}

	// Prevent unneccesary k8s updates
	var needsK8sUpdate bool = false
	if params.Name != nil && *params.Name != claim.Name || params.Zone != nil && *params.Zone != claim.Zone {
		needsK8sUpdate = true
	} else if params.Requested != nil && requestDiff(*params.Requested, claim.Requested) {
		needsK8sUpdate = true
	}

	if needsK8sUpdate {
		// TODO: impl update for k8s
	}

	return nil
}

// requestDiff checks if there are diffs in the requests for gpus
// Used to determine if k8s update is neccesary or if its fine to omit it and
// just update the db state
func requestDiff(newReqs []model.RequestedGpuCreate, oldReqs map[string]model.RequestedGpu) bool {
	if len(newReqs) != len(oldReqs) {
		return true
	}
	for i := range newReqs {
		key := newReqs[i].Name
		if _, found := oldReqs[key]; !found {
			return true
		}

		if oldReqs[key].AllocationMode != newReqs[i].AllocationMode {
			return true
		}

		if newReqs[i].AllocationMode == model.RequestAllocationMode_ExactCount {
			if oldReqs[key].Count != nil && newReqs[i].Count != nil {
				if *oldReqs[key].Count != *newReqs[i].Count {
					return true
				}
			} else if oldReqs[key].Count != newReqs[i].Count { // one is nil so we just check if both are
				return true
			}
		}

		if oldReqs[key].DeviceClassName != newReqs[i].DeviceClassName {
			return true
		}

		if len(oldReqs[key].Selectors) != len(newReqs[i].Selectors) {
			return true
		} else {
			for j := range oldReqs[key].Selectors {
				if oldReqs[key].Selectors[j] != newReqs[i].Selectors[j] {
					return true
				}
			}
		}

		// Check if ther is a diff in the Capacity
		if len(oldReqs[key].Capacity) != len(newReqs[i].Capacity) {
			return true
		}
		for k, v := range newReqs[i].Capacity {
			if vv, exists := oldReqs[key].Capacity[k]; exists {
				if v != vv {
					return true
				}
			} else {
				return true
			}
		}

		// Check if there is a diff in the config
		if oldReqs[key].Config != nil && newReqs[i].Config != nil {
			oldCfg := *oldReqs[key].Config
			newCfg := *newReqs[i].Config
			if oldCfg.Driver != newCfg.Driver {
				return true
			}
			if oldCfg.Parameters != nil && newCfg.Parameters != nil {
				if oldNvParams, ok := oldCfg.Parameters.(nvidia.GPUConfigParametersImpl); ok {
					if newNvParams, ok := newCfg.Parameters.(nvidia.GPUConfigParametersImpl); ok {

						if oldNvParams.APIVersion != newNvParams.APIVersion || oldNvParams.Kind != newNvParams.Kind {
							return true
						}
						if oldNvParams.Sharing != nil && newNvParams.Sharing != nil {
							oldSharing := *oldNvParams.Sharing
							newSharing := *newNvParams.Sharing

							if newSharing.Strategy != oldSharing.Strategy {
								return true
							}

							if oldSharing.MpsConfig != nil && newSharing.MpsConfig != nil {
								if oldSharing.MpsConfig.DefaultActiveThreadPercentage != nil && newSharing.MpsConfig.DefaultActiveThreadPercentage != nil {
									if *oldSharing.MpsConfig.DefaultActiveThreadPercentage != *newSharing.MpsConfig.DefaultActiveThreadPercentage {
										return true
									}
								} else if oldSharing.MpsConfig.DefaultActiveThreadPercentage != newSharing.MpsConfig.DefaultActiveThreadPercentage { // one is nil so we just check if both are
									return true
								}
								if len(oldSharing.MpsConfig.DefaultPerDevicePinnedMemoryLimit) != len(newSharing.MpsConfig.DefaultPerDevicePinnedMemoryLimit) {
									return true
								} else {
									for k, v := range newSharing.MpsConfig.DefaultPerDevicePinnedMemoryLimit {
										if vv, exists := oldSharing.MpsConfig.DefaultPerDevicePinnedMemoryLimit[k]; exists {
											if v.String() != vv.String() {
												return true
											}
										} else {
											return true
										}
									}
								}
								if oldSharing.MpsConfig.DefaultPinnedDeviceMemoryLimit != nil && newSharing.MpsConfig.DefaultPinnedDeviceMemoryLimit != nil {
									if oldSharing.MpsConfig.DefaultPinnedDeviceMemoryLimit.String() != newSharing.MpsConfig.DefaultPinnedDeviceMemoryLimit.String() {
										return true
									}
								} else if oldSharing.MpsConfig.DefaultPinnedDeviceMemoryLimit != newSharing.MpsConfig.DefaultPinnedDeviceMemoryLimit { // one is nil so we just check if both are
									return true
								}
							} else if oldSharing.MpsConfig != newSharing.MpsConfig { // one is nil so we just check if both are
								return true
							}
							if oldSharing.TimeSlicingConfig != nil && newSharing.TimeSlicingConfig != nil {
								if oldSharing.TimeSlicingConfig.Interval != nil && newSharing.TimeSlicingConfig != nil {
									if oldSharing.TimeSlicingConfig.Interval != nil && newSharing.TimeSlicingConfig.Interval != nil {
										if *oldSharing.TimeSlicingConfig.Interval != *newSharing.TimeSlicingConfig.Interval {
											return true
										}
									} else if oldSharing.TimeSlicingConfig.Interval != newSharing.TimeSlicingConfig.Interval { // one is nil so we just check if both are
										return true
									}
								} else if oldSharing.TimeSlicingConfig.Interval != newSharing.TimeSlicingConfig.Interval { // one is nil so we just check if both are
									return true
								}
							} else if oldSharing.TimeSlicingConfig != newSharing.TimeSlicingConfig { // one is nil so we just check if both are
								return true
							}
						} else if oldNvParams.Sharing != newNvParams.Sharing { // one is nil so we just check if both are
							return true
						}
					} else {
						return true // one is nvidia the other is not
					}
				} else if oldCfg.Parameters.MetaAPIVersion() != newCfg.Parameters.MetaAPIVersion() || oldCfg.Parameters.MetaKind() != newCfg.Parameters.MetaKind() {
					return true
				}
			} else if oldCfg.Parameters != newCfg.Parameters { // one is nil so we just check if both are
				return true
			}
		} else if oldReqs[key].Config != newReqs[i].Config { // one is nil so we just check if both are
			return true
		}
	}
	return false
}

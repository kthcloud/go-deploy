package v2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db/resources/gpu_claim_repo"
	jErrors "github.com/kthcloud/go-deploy/pkg/jobs/errors"
	"github.com/kthcloud/go-deploy/pkg/jobs/utils"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/parsers"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/parsers/dra"
	"github.com/kthcloud/go-deploy/service"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
)

func CreateGpuClaim(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	paramsMap, ok := job.Args["params"].(map[string]any)
	if !ok {
		return jErrors.MakeTerminatedError(fmt.Errorf("invalid params type"))
	}

	params, err := DecodeGpuClaimCreateParams(paramsMap)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = service.V2(utils.GetAuthInfo(job)).GpuClaims().Create(id, &params)
	if err != nil {
		// We always terminate these jobs, since rerunning it would cause a ErrNonUniqueField
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}

func DeleteGpuClaim(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	err = gpu_claim_repo.New().AddActivity(id, model.ActivityBeingDeleted)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	err = service.V2(utils.GetAuthInfo(job)).GpuClaims().Delete(id)
	if err != nil {
		if !errors.Is(err, sErrors.ErrResourceNotFound) {
			return jErrors.MakeFailedError(err)
		}
	}

	return nil
}

func UpdateGpuClaim(job *model.Job) error {
	err := utils.AssertParameters(job, []string{"id", "params"})
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	id := job.Args["id"].(string)

	paramsMap, ok := job.Args["params"].(map[string]any)
	if !ok {
		return jErrors.MakeTerminatedError(fmt.Errorf("invalid params type"))
	}

	params, err := DecodeGpuClaimUpdateParams(paramsMap)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}
	if err := service.V2(utils.GetAuthInfo(job)).GpuClaims().Update(id, &params); err != nil {
		if errors.Is(err, sErrors.ErrResourceNotFound) {
			return jErrors.MakeTerminatedError(err)
		}
	}
	return nil
}

// DecodeGpuClaimCreateParams decodes mapstructure for gpuClaimCreate.
// It is used because the gpuClaimCreateParams have such a complex format.
func DecodeGpuClaimCreateParams(raw map[string]any) (model.GpuClaimCreateParams, error) {
	var result model.GpuClaimCreateParams

	type tempGpuDeviceConfiguration struct {
		Driver     string `json:"driver"`
		Parameters any    `json:"parameters"`
	}
	type tempRequestedGpu struct {
		AllocationMode  string                      `json:"allocationMode"`
		DeviceClassName string                      `json:"deviceClassName"`
		Name            string                      `json:"name"`
		Config          *tempGpuDeviceConfiguration `json:"config,omitempty"`
	}
	type tempGpuClaimCreateParams struct {
		Name         string             `json:"name"`
		Zone         string             `json:"zone"`
		Requested    []tempRequestedGpu `json:"requested"`
		AllowedRoles []string           `json:"allowedRoles,omitempty"`
	}

	var temp tempGpuClaimCreateParams

	data, err := json.Marshal(raw)
	if err != nil {
		return result, fmt.Errorf("failed to marshal raw input: %w", err)
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return result, fmt.Errorf("failed to unmarshal into temp struct: %w", err)
	}

	result.Name = temp.Name
	result.Zone = temp.Zone
	result.AllowedRoles = temp.AllowedRoles
	result.Requested = make([]model.RequestedGpuCreate, len(temp.Requested))

	for i, r := range temp.Requested {
		req := model.RequestedGpuCreate{
			RequestedGpu: model.RequestedGpu{
				AllocationMode:  model.RequestAllocationMode(r.AllocationMode),
				DeviceClassName: r.DeviceClassName,
			},

			Name: r.Name,
		}

		if r.Config != nil {
			paramsJSON, err := json.Marshal(r.Config.Parameters)
			if err != nil {
				return result, fmt.Errorf("failed to marshal Parameters: %w", err)
			}

			opaqueParams, err := parsers.Parse[dra.OpaqueParams](bytes.NewReader(paramsJSON))
			if err != nil {
				return result, fmt.Errorf("failed to parse Parameters: %w", err)
			}

			req.Config = &model.GpuDeviceConfiguration{
				Driver:     r.Config.Driver,
				Parameters: opaqueParams,
			}
		}

		result.Requested[i] = req
	}

	return result, nil
}

// DecodeGpuClaimUpdateParams decodes mapstructure for gpuClaimUpdate.
// It is used because the gpuClaimUpdateParams have such a complex format.
func DecodeGpuClaimUpdateParams(raw map[string]any) (model.GpuClaimUpdateParams, error) {
	var result model.GpuClaimUpdateParams

	type tempGpuDeviceConfiguration struct {
		Driver     string `json:"driver"`
		Parameters any    `json:"parameters"`
	}
	type tempRequestedGpu struct {
		AllocationMode  string                      `json:"allocationMode"`
		DeviceClassName string                      `json:"deviceClassName"`
		Name            string                      `json:"name"`
		Config          *tempGpuDeviceConfiguration `json:"config,omitempty"`
	}
	type tempGpuClaimCreateParams struct {
		Name         *string             `json:"name,omitempty"`
		Zone         *string             `json:"zone,omitempty"`
		Requested    *[]tempRequestedGpu `json:"requested,omitempty"`
		AllowedRoles *[]string           `json:"allowedRoles,omitempty"`
	}

	var temp tempGpuClaimCreateParams

	data, err := json.Marshal(raw)
	if err != nil {
		return result, fmt.Errorf("failed to marshal raw input: %w", err)
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return result, fmt.Errorf("failed to unmarshal into temp struct: %w", err)
	}

	result.Name = temp.Name
	result.Zone = temp.Zone
	result.AllowedRoles = temp.AllowedRoles
	if temp.Requested != nil {
		resmap := make([]model.RequestedGpuCreate, len(*temp.Requested))
		result.Requested = &resmap

		for i, r := range *temp.Requested {
			req := model.RequestedGpuCreate{
				RequestedGpu: model.RequestedGpu{
					AllocationMode:  model.RequestAllocationMode(r.AllocationMode),
					DeviceClassName: r.DeviceClassName,
				},

				Name: r.Name,
			}

			if r.Config != nil {
				paramsJSON, err := json.Marshal(r.Config.Parameters)
				if err != nil {
					return result, fmt.Errorf("failed to marshal Parameters: %w", err)
				}

				opaqueParams, err := parsers.Parse[dra.OpaqueParams](bytes.NewReader(paramsJSON))
				if err != nil {
					return result, fmt.Errorf("failed to parse Parameters: %w", err)
				}

				req.Config = &model.GpuDeviceConfiguration{
					Driver:     r.Config.Driver,
					Parameters: opaqueParams,
				}
			}

			resmap[i] = req
		}
	}

	return result, nil
}

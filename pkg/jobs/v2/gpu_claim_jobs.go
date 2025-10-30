package v2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"maps"

	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db/resources/gpu_claim_repo"
	jErrors "github.com/kthcloud/go-deploy/pkg/jobs/errors"
	"github.com/kthcloud/go-deploy/pkg/jobs/utils"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/parsers"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/parsers/dra"
	"github.com/kthcloud/go-deploy/service"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
	"github.com/mitchellh/mapstructure"
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

	if pretty, err := json.MarshalIndent(params, "", "  "); err == nil {
		fmt.Println("bound:" + string(pretty))
	} else {
		fmt.Println("failed to pretty-print:", err)
	}

	if pretty, err := json.MarshalIndent(job.Args["params"].(map[string]any), "", "  "); err == nil {
		fmt.Println("raw:" + string(pretty))
	} else {
		fmt.Println("failed to pretty-print:", err)
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

func DecodeGpuClaimCreateParams(raw map[string]any) (model.GpuClaimCreateParams, error) {
	var result model.GpuClaimCreateParams

	// Make a shallow copy and remove Config from Requested items
	rawCopy := make(map[string]any)
	maps.Copy(rawCopy, raw)

	requestedRaw, _ := rawCopy["Requested"]
	requestedSlice, ok := requestedRaw.([]any)
	if !ok || requestedSlice == nil {
		requestedSlice = []any{}
	}

	// Remove Config from map so mapstructure won't try to decode it
	for _, r := range requestedSlice {
		if reqMap, ok := r.(map[string]any); ok {
			delete(reqMap, "Config")
		}
	}
	rawCopy["Requested"] = requestedSlice

	// Decode top-level fields + Requested without Config
	if err := mapstructure.Decode(rawCopy, &result); err != nil {
		return result, fmt.Errorf("failed to decode top-level fields: %w", err)
	}

	// manually parse Config.Parameters and assign
	for i, r := range requestedSlice {
		reqMap, ok := r.(map[string]any)
		if !ok {
			continue
		}

		configRaw, exists := reqMap["Config"]
		if !exists || configRaw == nil {
			continue
		}

		cfgMap, ok := configRaw.(map[string]any)
		if !ok {
			continue
		}

		rawParams, err := json.Marshal(cfgMap["Parameters"])
		if err != nil {
			return result, fmt.Errorf("failed to marshal Parameters: %w", err)
		}

		opaqueParams, err := parsers.Parse[dra.OpaqueParams](bytes.NewReader(rawParams))
		if err != nil {
			return result, fmt.Errorf("failed to parse Parameters: %w", err)
		}

		// put back parsed Config into the result struct
		result.Requested[i].Config = &model.GpuDeviceConfiguration{
			Driver:     cfgMap["driver"].(string),
			Parameters: opaqueParams,
		}
	}

	return result, nil
}

package v2

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

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

func CreateGpuClaim(ctx context.Context, job *model.Job) error {
	id, paramsMap, err := extractIDAndParams(job)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	params, err := DecodeGpuClaimCreateParams(paramsMap)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	if err := service.V2(utils.GetAuthInfo(job)).GpuClaims().Create(ctx, id, &params); err != nil {
		// We always terminate these jobs, since rerunning it would cause a ErrNonUniqueField
		return jErrors.MakeTerminatedError(err)
	}

	return nil
}

func DeleteGpuClaim(ctx context.Context, job *model.Job) error {
	id, err := extractJobID(job)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	if err := gpu_claim_repo.New().AddActivity(id, model.ActivityBeingDeleted); err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	if err := service.V2(utils.GetAuthInfo(job)).GpuClaims().Delete(ctx, id); err != nil {
		if !errors.Is(err, sErrors.ErrResourceNotFound) {
			return jErrors.MakeFailedError(err)
		}
	}

	return nil
}

func UpdateGpuClaim(ctx context.Context, job *model.Job) error {
	id, paramsMap, err := extractIDAndParams(job)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	params, err := DecodeGpuClaimUpdateParams(paramsMap)
	if err != nil {
		return jErrors.MakeTerminatedError(err)
	}

	if err := service.V2(utils.GetAuthInfo(job)).GpuClaims().Update(ctx, id, &params); err != nil {
		if errors.Is(err, sErrors.ErrResourceNotFound) {
			return jErrors.MakeTerminatedError(err)
		}
		return jErrors.MakeFailedError(err)
	}

	return nil
}

// DecodeGpuClaimCreateParams decodes mapstructure for gpuClaimCreate.
func DecodeGpuClaimCreateParams(raw map[string]any) (model.GpuClaimCreateParams, error) {
	return decodeGpuClaimParams[model.GpuClaimCreateParams](raw)
}

// DecodeGpuClaimUpdateParams decodes mapstructure for gpuClaimUpdate.
func DecodeGpuClaimUpdateParams(raw map[string]any) (model.GpuClaimUpdateParams, error) {
	return decodeGpuClaimParams[model.GpuClaimUpdateParams](raw)
}

func decodeGpuClaimParams[T any](raw map[string]any) (T, error) {
	var result T
	if raw == nil {
		return result, fmt.Errorf("raw params is nil")
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: opaqueParamsDecodeHook(),
		Result:     &result,
	})
	if err != nil {
		return result, fmt.Errorf("failed to create decoder: %w", err)
	}

	if err := decoder.Decode(raw); err != nil {
		return result, fmt.Errorf("failed to decode params: %w", err)
	}

	return result, nil
}

func opaqueParamsDecodeHook() mapstructure.DecodeHookFunc {
	opaqueType := reflect.TypeFor[dra.OpaqueParams]()
	return func(_ reflect.Type, t reflect.Type, data any) (any, error) {
		if t != opaqueType || data == nil {
			return data, nil
		}
		if p, ok := data.(dra.OpaqueParams); ok {
			return p, nil
		}

		rawBytes, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal opaque parameters: %w", err)
		}

		trimmed := bytes.TrimSpace(rawBytes)
		if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
			return nil, nil
		}

		opaqueParams, err := parsers.Parse[dra.OpaqueParams](bytes.NewReader(trimmed))
		if err != nil {
			return nil, fmt.Errorf("failed to parse opaque parameters: %w", err)
		}

		return opaqueParams, nil
	}
}

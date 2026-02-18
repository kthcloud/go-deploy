package nvidia

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/api/nvidia"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/parsers/dra"
)

const (
	supportedApiVersion = "resource.nvidia.com/v1beta1"
	supportedKind       = "GpuConfig"
)

type GPUConfigParametersImpl struct {
	nvidia.GpuConfig
}

func (s GPUConfigParametersImpl) MetaAPIVersion() string {
	return s.APIVersion
}

func (s GPUConfigParametersImpl) MetaKind() string {
	return s.Kind
}

type GPUConfigParametersParserImpl struct {
}

func (GPUConfigParametersParserImpl) CanParse(raw io.Reader) bool {
	decoder := json.NewDecoder(raw)
	var tmp map[string]any
	if err := decoder.Decode(&tmp); err != nil {
		return false
	}

	// Helper to find a key ignoring case
	findKey := func(m map[string]any, key string) (any, bool) {
		for k, v := range m {
			if strings.EqualFold(k, key) {
				return v, true
			}
		}
		return nil, false
	}

	// Get apiVersion ignoring key case
	if apiVersionRaw, ok := findKey(tmp, "apiVersion"); ok {
		if apiVersion, ok := apiVersionRaw.(string); ok && strings.EqualFold(apiVersion, supportedApiVersion) {
			if kindRaw, ok := findKey(tmp, "kind"); ok {
				if kind, ok := kindRaw.(string); ok && strings.EqualFold(kind, supportedKind) {
					return true
				}
			}
		}
	}

	return false
}

func (GPUConfigParametersParserImpl) Parse(raw io.Reader) (dra.OpaqueParams, error) {
	var gpuCfg nvidia.GpuConfig
	decoder := json.NewDecoder(raw)
	if err := decoder.Decode(&gpuCfg); err != nil {
		return nil, err
	}

	return GPUConfigParametersImpl{GpuConfig: gpuCfg}, nil
}

package models

import (
	"encoding/json"
	"time"

	nvresourcebetav1 "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	resourcev1 "k8s.io/api/resource/v1"
)

type ResourceClaimTemplatePublic struct {
	Name             string    `bson:"name"`
	Namespace        string    `bson:"namespace"`
	DeviceClass      string    `bson:"deviceClass"`
	Requests         []string  `bson:"requests"`
	Driver           string    `bson:"driver"`
	Strategy         string    `bson:"strategy"`
	MPSActiveThreads int       `bson:"mpsActiveThreads"`
	MPSMemoryLimit   string    `bson:"mpsMemoryLimit"`
	CreatedAt        time.Time `bson:"createdAt"`
}

func (r *ResourceClaimTemplatePublic) Created() bool {
	return r.CreatedAt != time.Time{}
}

func (r *ResourceClaimTemplatePublic) IsPlaceholder() bool {
	return false
}

func CreateResourceClaimTemplatePublicFromRead(tmpl *resourcev1.ResourceClaimTemplate) *ResourceClaimTemplatePublic {
	var deviceClass, driver, strategy string
	var requests []string
	var mpsActiveThreads int
	var mpsMemoryLimit string

	if len(tmpl.Spec.Spec.Devices.Requests) > 0 {
		req := tmpl.Spec.Spec.Devices.Requests[0]
		if req.Exactly != nil {
			deviceClass = req.Exactly.DeviceClassName
		}
	}

	if len(tmpl.Spec.Spec.Devices.Config) > 0 {
		cfg := tmpl.Spec.Spec.Devices.Config[0]
		requests = cfg.Requests
		if cfg.Opaque.Driver != "" {
			driver = cfg.Opaque.Driver
		}

		// Decode NVIDIA-specific GPU config if available
		if len(cfg.Opaque.Parameters.Raw) > 0 {
			var gpuCfg nvresourcebetav1.GpuConfig
			if err := json.Unmarshal(cfg.Opaque.Parameters.Raw, &gpuCfg); err == nil {
				strategy = string(gpuCfg.Sharing.Strategy)
				if gpuCfg.Sharing.MpsConfig != nil {
					mpsActiveThreads = *gpuCfg.Sharing.MpsConfig.DefaultActiveThreadPercentage
					mpsMemoryLimit = gpuCfg.Sharing.MpsConfig.DefaultPinnedDeviceMemoryLimit.String()
				}
			}
		}
	}

	return &ResourceClaimTemplatePublic{
		Name:             tmpl.Name,
		Namespace:        tmpl.Namespace,
		DeviceClass:      deviceClass,
		Requests:         requests,
		Driver:           driver,
		Strategy:         strategy,
		MPSActiveThreads: mpsActiveThreads,
		MPSMemoryLimit:   mpsMemoryLimit,
		CreatedAt:        formatCreatedAt(tmpl.Annotations),
	}
}

package models

import (
	"go-deploy/pkg/subsystems/k8s/keys"
	v2 "k8s.io/api/autoscaling/v2"
	apiv1 "k8s.io/api/core/v1"
	"time"
)

type HpaPublic struct {
	ID                       string    `bson:"id"`
	Name                     string    `bson:"name"`
	Namespace                string    `bson:"namespace"`
	MinReplicas              int       `bson:"minReplicas"`
	MaxReplicas              int       `bson:"maxReplicas"`
	Target                   string    `bson:"target"`
	TargetKind               string    `bson:"targetKind"`
	CpuAverageUtilization    int       `bson:"cpuAverageUtilization"`
	MemoryAverageUtilization int       `bson:"memoryAverageUtilization"`
	CreatedAt                time.Time `bson:"createdAt"`
}

func (h *HpaPublic) GetID() string {
	return h.ID
}

func (h *HpaPublic) Created() bool {
	return h.ID != ""
}

func (h *HpaPublic) IsPlaceholder() bool {
	return false
}

func CreateHpaPublicFromRead(hpa *v2.HorizontalPodAutoscaler) *HpaPublic {
	var minReplicas int
	var maxReplicas int
	var cpuAverageUtilization int
	var memoryAverageUtilization int

	if hpa.Spec.MinReplicas != nil {
		minReplicas = int(*hpa.Spec.MinReplicas)
	}

	if hpa.Spec.MaxReplicas != 0 {
		maxReplicas = int(hpa.Spec.MaxReplicas)
	}

	if hpa.Spec.Metrics != nil && len(hpa.Spec.Metrics) > 0 {
		for _, metric := range hpa.Spec.Metrics {
			if metric.Resource != nil && metric.Resource.Name == apiv1.ResourceCPU && metric.Resource.Target.AverageValue != nil {
				cpuAverageUtilization = int(*metric.Resource.Target.AverageUtilization)
			}

			if metric.Resource != nil && metric.Resource.Name == apiv1.ResourceMemory && metric.Resource.Target.AverageUtilization != nil {
				memoryAverageUtilization = int(*metric.Resource.Target.AverageUtilization)
			}
		}
	}

	return &HpaPublic{
		ID:                       hpa.Labels[keys.ManifestLabelID],
		Name:                     hpa.Name,
		Namespace:                hpa.Namespace,
		MinReplicas:              minReplicas,
		MaxReplicas:              maxReplicas,
		Target:                   hpa.Spec.ScaleTargetRef.Name,
		TargetKind:               hpa.Spec.ScaleTargetRef.Kind,
		CpuAverageUtilization:    cpuAverageUtilization,
		MemoryAverageUtilization: memoryAverageUtilization,
		CreatedAt:                formatCreatedAt(hpa.Annotations),
	}
}

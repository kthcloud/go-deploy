package model

import k8sModels "github.com/kthcloud/go-deploy/pkg/subsystems/k8s/models"

type GpuClaimSubsystems struct {
	K8s GpuClaimK8s `bson:"k8s"`
}

type GpuClaimK8s struct {
	Namespace k8sModels.NamespacePublic `bson:"namespace"`

	ResourceClaimMap map[string]k8sModels.ResourceClaimPublic `bson:"resourceClaimMap,omitempty"`
}

// GetNamespace returns the namespace of the resourceClaim.
func (k *GpuClaimK8s) GetNamespace() *k8sModels.NamespacePublic {
	return &k.Namespace
}

package model

import (
	harborModels "github.com/kthcloud/go-deploy/pkg/subsystems/harbor/models"
	k8sModels "github.com/kthcloud/go-deploy/pkg/subsystems/k8s/models"
)

type DeploymentSubsystems struct {
	K8s    DeploymentK8s    `bson:"k8s"`
	Harbor DeploymentHarbor `bson:"harbor"`
}

type DeploymentK8s struct {
	Namespace k8sModels.NamespacePublic `bson:"namespace"`

	DeploymentMap    map[string]k8sModels.DeploymentPublic            `bson:"deploymentMap,omitempty"`
	ServiceMap       map[string]k8sModels.ServicePublic               `bson:"serviceMap,omitempty"`
	IngressMap       map[string]k8sModels.IngressPublic               `bson:"ingressMap,omitempty"`
	PvMap            map[string]k8sModels.PvPublic                    `bson:"pvMap,omitempty"`
	PvcMap           map[string]k8sModels.PvcPublic                   `bson:"pvcMap,omitempty"`
	RCTMap           map[string]k8sModels.ResourceClaimTemplatePublic `bson:"rctMap,omitempty"`
	SecretMap        map[string]k8sModels.SecretPublic                `bson:"secretMap,omitempty"`
	HpaMap           map[string]k8sModels.HpaPublic                   `bson:"hpaMap,omitempty"`
	NetworkPolicyMap map[string]k8sModels.NetworkPolicyPublic         `bson:"networkPolicyMap,omitempty"`
}

type DeploymentHarbor struct {
	Project     harborModels.ProjectPublic    `bson:"project"`
	Robot       harborModels.RobotPublic      `bson:"robot"`
	Repository  harborModels.RepositoryPublic `bson:"repository"`
	Webhook     harborModels.WebhookPublic    `bson:"webhook"`
	Placeholder bool                          `bson:"placeholder"`
}

// GetNamespace returns the namespace of the deployment.
func (k *DeploymentK8s) GetNamespace() *k8sModels.NamespacePublic {
	return &k.Namespace
}

// GetDeploymentMap returns the deployment map of the deployment.
// If the map is nil, it will be initialized before returning.
func (k *DeploymentK8s) GetDeploymentMap() map[string]k8sModels.DeploymentPublic {
	if k.DeploymentMap == nil {
		k.DeploymentMap = make(map[string]k8sModels.DeploymentPublic)
	}

	return k.DeploymentMap
}

// GetServiceMap returns the service map of the deployment.
// If the map is nil, it will be initialized before returning.
func (k *DeploymentK8s) GetServiceMap() map[string]k8sModels.ServicePublic {
	if k.ServiceMap == nil {
		k.ServiceMap = make(map[string]k8sModels.ServicePublic)
	}

	return k.ServiceMap
}

// GetIngressMap returns the ingress map of the deployment.
// If the map is nil, it will be initialized before returning.
func (k *DeploymentK8s) GetIngressMap() map[string]k8sModels.IngressPublic {
	if k.IngressMap == nil {
		k.IngressMap = make(map[string]k8sModels.IngressPublic)
	}

	return k.IngressMap
}

// GetPvMap returns the pv map of the deployment.
// If the map is nil, it will be initialized before returning.
func (k *DeploymentK8s) GetPvMap() map[string]k8sModels.PvPublic {
	if k.PvMap == nil {
		k.PvMap = make(map[string]k8sModels.PvPublic)
	}

	return k.PvMap
}

// GetPvcMap returns the pvc map of the deployment.
// If the map is nil, it will be initialized before returning.
func (k *DeploymentK8s) GetPvcMap() map[string]k8sModels.PvcPublic {
	if k.PvcMap == nil {
		k.PvcMap = make(map[string]k8sModels.PvcPublic)
	}

	return k.PvcMap
}

// GetRCTMap returns the resourceClaimTemplate map of the deployment.
// If the map is nil, it will be initialized before returning.
func (k *DeploymentK8s) GetRCTMap() map[string]k8sModels.ResourceClaimTemplatePublic {
	if k.RCTMap == nil {
		k.RCTMap = make(map[string]k8sModels.ResourceClaimTemplatePublic)
	}

	return k.RCTMap
}

// GetSecretMap returns the secret map of the deployment.
// If the map is nil, it will be initialized before returning.
func (k *DeploymentK8s) GetSecretMap() map[string]k8sModels.SecretPublic {
	if k.SecretMap == nil {
		k.SecretMap = make(map[string]k8sModels.SecretPublic)
	}

	return k.SecretMap
}

// GetHpaMap returns the hpa map of the deployment.
// If the map is nil, it will be initialized before returning.
func (k *DeploymentK8s) GetHpaMap() map[string]k8sModels.HpaPublic {
	if k.HpaMap == nil {
		k.HpaMap = make(map[string]k8sModels.HpaPublic)
	}

	return k.HpaMap
}

// GetNetworkPolicyMap returns the network policy map of the deployment.
// If the map is nil, it will be initialized before returning.
func (k *DeploymentK8s) GetNetworkPolicyMap() map[string]k8sModels.NetworkPolicyPublic {
	if k.NetworkPolicyMap == nil {
		k.NetworkPolicyMap = make(map[string]k8sModels.NetworkPolicyPublic)
	}

	return k.NetworkPolicyMap
}

// GetDeployment returns the deployment with the given name.
// If a deployment with the given name does not exist, nil will be returned.
func (k *DeploymentK8s) GetDeployment(name string) *k8sModels.DeploymentPublic {
	resource, ok := k.GetDeploymentMap()[name]
	if !ok {
		return nil
	}

	return &resource
}

// GetService returns the service with the given name.
// If a service with the given name does not exist, nil will be returned.
func (k *DeploymentK8s) GetService(name string) *k8sModels.ServicePublic {
	resource, ok := k.GetServiceMap()[name]
	if !ok {
		return nil
	}

	return &resource
}

// GetIngress returns the ingress with the given name.
// If an ingress with the given name does not exist, nil will be returned.
func (k *DeploymentK8s) GetIngress(name string) *k8sModels.IngressPublic {
	resource, ok := k.GetIngressMap()[name]
	if !ok {
		return nil
	}

	return &resource
}

// GetPV returns the PV with the given name.
// If a PV with the given name does not exist, nil will be returned.
func (k *DeploymentK8s) GetPV(name string) *k8sModels.PvPublic {
	resource, ok := k.GetPvMap()[name]
	if !ok {
		return nil
	}

	return &resource
}

// GetPVC returns the PVC with the given name.
// If a PVC with the given name does not exist, nil will be returned.
func (k *DeploymentK8s) GetPVC(name string) *k8sModels.PvcPublic {
	resource, ok := k.GetPvcMap()[name]
	if !ok {
		return nil
	}

	return &resource
}

// GetRCT returns the ResourceClaimTemplate with the given name.
// If a RCT with the given name does not exist, nil will be returned.
func (k *DeploymentK8s) GetRCT(name string) *k8sModels.ResourceClaimTemplatePublic {
	resource, ok := k.GetRCTMap()[name]
	if !ok {
		return nil
	}

	return &resource
}

// GetSecret returns the secret with the given name.
// If a secret with the given name does not exist, nil will be returned.
func (k *DeploymentK8s) GetSecret(name string) *k8sModels.SecretPublic {
	resource, ok := k.GetSecretMap()[name]
	if !ok {
		return nil
	}

	return &resource
}

// GetHPA returns the HPA with the given name.
// If a HPA with the given name does not exist, nil will be returned.
func (k *DeploymentK8s) GetHPA(name string) *k8sModels.HpaPublic {
	resource, ok := k.GetHpaMap()[name]
	if !ok {
		return nil
	}

	return &resource
}

// GetNetworkPolicy returns the network policy with the given name.
// If a network policy with the given name does not exist, nil will be returned.
func (k *DeploymentK8s) GetNetworkPolicy(name string) *k8sModels.NetworkPolicyPublic {
	resource, ok := k.GetNetworkPolicyMap()[name]
	if !ok {
		return nil
	}

	return &resource
}

// DeleteDeployment deletes the deployment from the DeploymentMap with the given name.
// It uses GetDeploymentMap to get the map to avoid nil pointer dereferences.
func (k *DeploymentK8s) DeleteDeployment(name string) {
	delete(k.GetDeploymentMap(), name)
}

// DeleteService deletes the service from the ServiceMap with the given name.
// It uses GetServiceMap to get the map to avoid nil pointer dereferences.
func (k *DeploymentK8s) DeleteService(name string) {
	delete(k.GetServiceMap(), name)
}

// DeleteIngress deletes the ingress from the IngressMap with the given name.
// It uses GetIngressMap to get the map to avoid nil pointer dereferences.
func (k *DeploymentK8s) DeleteIngress(name string) {
	delete(k.GetIngressMap(), name)
}

// DeletePV deletes the PV from the PvMap with the given name.
// It uses GetPvMap to get the map to avoid nil pointer dereferences.
func (k *DeploymentK8s) DeletePV(name string) {
	delete(k.GetPvMap(), name)
}

// DeletePVC deletes the PVC from the PvcMap with the given name.
// It uses GetPvcMap to get the map to avoid nil pointer dereferences.
func (k *DeploymentK8s) DeletePVC(name string) {
	delete(k.GetPvcMap(), name)
}

// DeleteRCT deletes the ResourceClaimTemplate from the RCTMap with the given name.
// It uses GetRCTMap to get the map to avoid nil pointer dereferences.
func (k *DeploymentK8s) DeleteRCT(name string) {
	delete(k.GetRCTMap(), name)
}

// DeleteSecret deletes the secret from the SecretMap with the given name.
// It uses GetSecretMap to get the map to avoid nil pointer dereferences.
func (k *DeploymentK8s) DeleteSecret(name string) {
	delete(k.GetSecretMap(), name)
}

// DeleteHPA deletes the HPA from the HpaMap with the given name.
// It uses GetHpaMap to get the map to avoid nil pointer dereferences.
func (k *DeploymentK8s) DeleteHPA(name string) {
	delete(k.GetHpaMap(), name)
}

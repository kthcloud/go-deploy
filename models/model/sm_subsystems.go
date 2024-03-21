package model

import (
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
)

type SmSubsystems struct {
	K8s    DeploymentK8s    `bson:"k8s"`
	Harbor DeploymentHarbor `bson:"harbor"`
}

type SmK8s struct {
	Namespace     k8sModels.NamespacePublic             `bson:"namespace"`
	DeploymentMap map[string]k8sModels.DeploymentPublic `bson:"deploymentMap,omitempty"`
	ServiceMap    map[string]k8sModels.ServicePublic    `bson:"serviceMap,omitempty"`
	IngressMap    map[string]k8sModels.IngressPublic    `bson:"ingressMap,omitempty"`
	PvMap         map[string]k8sModels.PvPublic         `bson:"pvMap,omitempty"`
	PvcMap        map[string]k8sModels.PvcPublic        `bson:"pvcMap,omitempty"`
}

// GetNamespace returns the namespace of the deployment.
func (k *SmK8s) GetNamespace() *k8sModels.NamespacePublic {
	return &k.Namespace
}

// GetDeploymentMap returns the deployment map of the deployment.
// If the map is nil, it will be initialized before returning.
func (k *SmK8s) GetDeploymentMap() map[string]k8sModels.DeploymentPublic {
	if k.DeploymentMap == nil {
		k.DeploymentMap = make(map[string]k8sModels.DeploymentPublic)
	}

	return k.DeploymentMap
}

// GetServiceMap returns the service map of the deployment.
// If the map is nil, it will be initialized before returning.
func (k *SmK8s) GetServiceMap() map[string]k8sModels.ServicePublic {
	if k.ServiceMap == nil {
		k.ServiceMap = make(map[string]k8sModels.ServicePublic)
	}

	return k.ServiceMap
}

// GetIngressMap returns the ingress map of the deployment.
// If the map is nil, it will be initialized before returning.
func (k *SmK8s) GetIngressMap() map[string]k8sModels.IngressPublic {
	if k.IngressMap == nil {
		k.IngressMap = make(map[string]k8sModels.IngressPublic)
	}

	return k.IngressMap
}

// GetPvMap returns the pv map of the deployment.
// If the map is nil, it will be initialized before returning.
func (k *SmK8s) GetPvMap() map[string]k8sModels.PvPublic {
	if k.PvMap == nil {
		k.PvMap = make(map[string]k8sModels.PvPublic)
	}

	return k.PvMap
}

// GetPvcMap returns the pvc map of the deployment.
// If the map is nil, it will be initialized before returning.
func (k *SmK8s) GetPvcMap() map[string]k8sModels.PvcPublic {
	if k.PvcMap == nil {
		k.PvcMap = make(map[string]k8sModels.PvcPublic)
	}

	return k.PvcMap
}

// GetDeployment returns the deployment with the given name.
// If a deployment with the given name does not exist, nil will be returned.
func (k *SmK8s) GetDeployment(name string) *k8sModels.DeploymentPublic {
	resource, ok := k.GetDeploymentMap()[name]
	if !ok {
		return nil
	}

	return &resource
}

// GetService returns the service with the given name.
// If a service with the given name does not exist, nil will be returned.
func (k *SmK8s) GetService(name string) *k8sModels.ServicePublic {
	resource, ok := k.GetServiceMap()[name]
	if !ok {
		return nil
	}

	return &resource
}

// GetIngress returns the ingress with the given name.
// If an ingress with the given name does not exist, nil will be returned.
func (k *SmK8s) GetIngress(name string) *k8sModels.IngressPublic {
	resource, ok := k.GetIngressMap()[name]
	if !ok {
		return nil
	}

	return &resource
}

// GetPV returns the PV with the given name.
// If a PV with the given name does not exist, nil will be returned.
func (k *SmK8s) GetPV(name string) *k8sModels.PvPublic {
	resource, ok := k.GetPvMap()[name]
	if !ok {
		return nil
	}

	return &resource
}

// GetPVC returns the PVC with the given name.
// If a PVC with the given name does not exist, nil will be returned.
func (k *SmK8s) GetPVC(name string) *k8sModels.PvcPublic {
	resource, ok := k.GetPvcMap()[name]
	if !ok {
		return nil
	}

	return &resource
}

// DeleteDeployment deletes the deployment from the DeploymentMap with the given name.
// It uses GetDeploymentMap to get the map to avoid nil pointer dereferences.
func (k *SmK8s) DeleteDeployment(name string) {
	delete(k.GetDeploymentMap(), name)
}

// DeleteService deletes the service from the ServiceMap with the given name.
// It uses GetServiceMap to get the map to avoid nil pointer dereferences.
func (k *SmK8s) DeleteService(name string) {
	delete(k.GetServiceMap(), name)
}

// DeleteIngress deletes the ingress from the IngressMap with the given name.
// It uses GetIngressMap to get the map to avoid nil pointer dereferences.
func (k *SmK8s) DeleteIngress(name string) {
	delete(k.GetIngressMap(), name)
}

// DeletePV deletes the PV from the PvMap with the given name.
// It uses GetPvMap to get the map to avoid nil pointer dereferences.
func (k *SmK8s) DeletePV(name string) {
	delete(k.GetPvMap(), name)
}

// DeletePVC deletes the PVC from the PvcMap with the given name.
// It uses GetPvcMap to get the map to avoid nil pointer dereferences.
func (k *SmK8s) DeletePVC(name string) {
	delete(k.GetPvcMap(), name)
}

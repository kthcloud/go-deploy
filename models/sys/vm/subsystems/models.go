package subsystems

import (
	csModels "go-deploy/pkg/subsystems/cs/models"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
)

type CS struct {
	ServiceOffering       csModels.ServiceOfferingPublic               `bson:"serviceOffering"`
	VM                    csModels.VmPublic                            `bson:"vm"`
	PortForwardingRuleMap map[string]csModels.PortForwardingRulePublic `bson:"portForwardingRuleMap,omitempty"`
	SnapshotMap           map[string]csModels.SnapshotPublic           `bson:"snapshotMap,omitempty"`
}

type K8s struct {
	Namespace     k8sModels.NamespacePublic             `bson:"namespace"`
	DeploymentMap map[string]k8sModels.DeploymentPublic `bson:"deploymentMap,omitempty"`
	ServiceMap    map[string]k8sModels.ServicePublic    `bson:"serviceMap,omitempty"`
	IngressMap    map[string]k8sModels.IngressPublic    `bson:"ingressMap,omitempty"`
	SecretMap     map[string]k8sModels.SecretPublic     `bson:"secretMap,omitempty"`
}

func (k8s *K8s) GetDeploymentMap() map[string]k8sModels.DeploymentPublic {
	if k8s.DeploymentMap == nil {
		k8s.DeploymentMap = make(map[string]k8sModels.DeploymentPublic)
	}

	return k8s.DeploymentMap
}

func (k8s *K8s) GetServiceMap() map[string]k8sModels.ServicePublic {
	if k8s.ServiceMap == nil {
		k8s.ServiceMap = make(map[string]k8sModels.ServicePublic)
	}

	return k8s.ServiceMap
}

func (k8s *K8s) GetIngressMap() map[string]k8sModels.IngressPublic {
	if k8s.IngressMap == nil {
		k8s.IngressMap = make(map[string]k8sModels.IngressPublic)
	}

	return k8s.IngressMap
}

func (k8s *K8s) GetDeployment(name string) *k8sModels.DeploymentPublic {
	resources, ok := k8s.GetDeploymentMap()[name]
	if !ok {
		return nil
	}

	return &resources
}

func (k8s *K8s) GetService(name string) *k8sModels.ServicePublic {
	resources, ok := k8s.GetServiceMap()[name]
	if !ok {
		return nil
	}

	return &resources
}

func (k8s *K8s) GetIngress(name string) *k8sModels.IngressPublic {
	resources, ok := k8s.GetIngressMap()[name]
	if !ok {
		return nil
	}

	return &resources
}

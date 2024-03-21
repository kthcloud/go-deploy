package model

import (
	csModels "go-deploy/pkg/subsystems/cs/models"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
)

// VmCS is only used in V1
type VmCS struct {
	VM                    csModels.VmPublic                            `bson:"vm"`
	PortForwardingRuleMap map[string]csModels.PortForwardingRulePublic `bson:"portForwardingRuleMap,omitempty"`
	SnapshotMap           map[string]csModels.SnapshotPublic           `bson:"snapshotMap,omitempty"`
}

type VmK8s struct {
	Namespace        k8sModels.NamespacePublic                `bson:"namespace"`
	DeploymentMap    map[string]k8sModels.DeploymentPublic    `bson:"deploymentMap,omitempty"`
	ServiceMap       map[string]k8sModels.ServicePublic       `bson:"serviceMap,omitempty"`
	IngressMap       map[string]k8sModels.IngressPublic       `bson:"ingressMap,omitempty"`
	SecretMap        map[string]k8sModels.SecretPublic        `bson:"secretMap,omitempty"`
	NetworkPolicyMap map[string]k8sModels.NetworkPolicyPublic `bson:"networkPolicyMap,omitempty"`

	// VmMap only exists in version 2
	VmMap map[string]k8sModels.VmPublic `bson:"vmMap,omitempty"`

	// PvcMap only exists in version 2
	PvcMap map[string]k8sModels.PvcPublic `bson:"pvcMap,omitempty"`

	// PvMap only exists in version 2
	PvMap map[string]k8sModels.PvPublic `bson:"pvMap,omitempty"`

	// SnapshotMap only exists in version 2
	VmSnapshotMap map[string]k8sModels.VmSnapshotPublic `bson:"vmSnapshotMap,omitempty"`
}

// GetDeploymentMap returns the deployment map of the VM.
// If the map is nil, it will be initialized before returning
func (k8s *VmK8s) GetDeploymentMap() map[string]k8sModels.DeploymentPublic {
	if k8s.DeploymentMap == nil {
		k8s.DeploymentMap = make(map[string]k8sModels.DeploymentPublic)
	}

	return k8s.DeploymentMap
}

// GetVmMap returns the vm map of the VM.
// If the map is nil, it will be initialized before returning
func (k8s *VmK8s) GetVmMap() map[string]k8sModels.VmPublic {
	if k8s.VmMap == nil {
		k8s.VmMap = make(map[string]k8sModels.VmPublic)
	}

	return k8s.VmMap
}

// GetVmSnapshotMap returns the vm snapshot map of the VM.
// If the map is nil, it will be initialized before returning
func (k8s *VmK8s) GetVmSnapshotMap() map[string]k8sModels.VmSnapshotPublic {
	if k8s.VmSnapshotMap == nil {
		k8s.VmSnapshotMap = make(map[string]k8sModels.VmSnapshotPublic)
	}

	return k8s.VmSnapshotMap
}

// GetServiceMap returns the service map of the VM.
// If the map is nil, it will be initialized before returning
func (k8s *VmK8s) GetServiceMap() map[string]k8sModels.ServicePublic {
	if k8s.ServiceMap == nil {
		k8s.ServiceMap = make(map[string]k8sModels.ServicePublic)
	}

	return k8s.ServiceMap
}

// GetIngressMap returns the ingress map of the VM.
// If the map is nil, it will be initialized before returning
func (k8s *VmK8s) GetIngressMap() map[string]k8sModels.IngressPublic {
	if k8s.IngressMap == nil {
		k8s.IngressMap = make(map[string]k8sModels.IngressPublic)
	}

	return k8s.IngressMap
}

// GetSecretMap returns the secret map of the VM.
// If the map is nil, it will be initialized before returning
func (k8s *VmK8s) GetSecretMap() map[string]k8sModels.SecretPublic {
	if k8s.SecretMap == nil {
		k8s.SecretMap = make(map[string]k8sModels.SecretPublic)
	}

	return k8s.SecretMap
}

// GetNetworkPolicyMap returns the network policy map of the VM.
// If the map is nil, it will be initialized before returning
func (k8s *VmK8s) GetNetworkPolicyMap() map[string]k8sModels.NetworkPolicyPublic {
	if k8s.NetworkPolicyMap == nil {
		k8s.NetworkPolicyMap = make(map[string]k8sModels.NetworkPolicyPublic)
	}

	return k8s.NetworkPolicyMap
}

// GetPvcMap returns the PVC map of the VM.
// If the map is nil, it will be initialized before returning
func (k8s *VmK8s) GetPvcMap() map[string]k8sModels.PvcPublic {
	if k8s.PvcMap == nil {
		k8s.PvcMap = make(map[string]k8sModels.PvcPublic)
	}

	return k8s.PvcMap
}

// GetPvMap returns the PV map of the VM.
// If the map is nil, it will be initialized before returning
func (k8s *VmK8s) GetPvMap() map[string]k8sModels.PvPublic {
	if k8s.PvMap == nil {
		k8s.PvMap = make(map[string]k8sModels.PvPublic)
	}

	return k8s.PvMap
}

// GetDeployment returns the deployment of the VM.
// If the deployment does not exist, nil is returned
func (k8s *VmK8s) GetDeployment(name string) *k8sModels.DeploymentPublic {
	resources, ok := k8s.GetDeploymentMap()[name]
	if !ok {
		return nil
	}

	return &resources
}

// GetVM returns the vm of the VM.
// If the vm does not exist, nil is returned
func (k8s *VmK8s) GetVM(name string) *k8sModels.VmPublic {
	resources, ok := k8s.GetVmMap()[name]
	if !ok {
		return nil
	}

	return &resources
}

// GetService returns the service of the VM.
// If the service does not exist, nil is returned
func (k8s *VmK8s) GetService(name string) *k8sModels.ServicePublic {
	resources, ok := k8s.GetServiceMap()[name]
	if !ok {
		return nil
	}

	return &resources
}

// GetIngress returns the ingress of the VM.
// If the ingress does not exist, nil is returned
func (k8s *VmK8s) GetIngress(name string) *k8sModels.IngressPublic {
	resources, ok := k8s.GetIngressMap()[name]
	if !ok {
		return nil
	}

	return &resources
}

// GetSecret returns the secret of the VM.
// If the secret does not exist, nil is returned
func (k8s *VmK8s) GetSecret(name string) *k8sModels.SecretPublic {
	resources, ok := k8s.GetSecretMap()[name]
	if !ok {
		return nil
	}

	return &resources
}

// GetNetworkPolicy returns the network policy of the VM.
// If the network policy does not exist, nil is returned
func (k8s *VmK8s) GetNetworkPolicy(name string) *k8sModels.NetworkPolicyPublic {
	resources, ok := k8s.GetNetworkPolicyMap()[name]
	if !ok {
		return nil
	}

	return &resources
}

// GetPVC returns the PVC of the VM.
// If the PVC does not exist, nil is returned
func (k8s *VmK8s) GetPVC(name string) *k8sModels.PvcPublic {
	resources, ok := k8s.GetPvcMap()[name]
	if !ok {
		return nil
	}

	return &resources
}

// GetPV returns the PV of the VM.
// If the PV does not exist, nil is returned
func (k8s *VmK8s) GetPV(name string) *k8sModels.PvPublic {
	resources, ok := k8s.GetPvMap()[name]
	if !ok {
		return nil
	}

	return &resources
}

// GetVmSnapshotByName returns a snapshot in the VM by name.
// If the snapshot does not exist, nil is returned
func (k8s *VmK8s) GetVmSnapshotByName(name string) *k8sModels.VmSnapshotPublic {
	resources, ok := k8s.GetVmSnapshotMap()[name]
	if !ok {
		return nil
	}

	return &resources
}

// GetVmSnapshotByID returns a snapshot in the VM by ID.
// If the snapshot does not exist, nil is returned
func (k8s *VmK8s) GetVmSnapshotByID(id string) *k8sModels.VmSnapshotPublic {
	for _, snapshot := range k8s.GetVmSnapshotMap() {
		if snapshot.ID == id {
			return &snapshot
		}
	}

	return nil
}

func (cs *VmCS) GetPortForwardingRuleMap() map[string]csModels.PortForwardingRulePublic {
	if cs.PortForwardingRuleMap == nil {
		cs.PortForwardingRuleMap = make(map[string]csModels.PortForwardingRulePublic)
	}

	return cs.PortForwardingRuleMap
}

func (cs *VmCS) GetSnapshotMap() map[string]csModels.SnapshotPublic {
	if cs.SnapshotMap == nil {
		cs.SnapshotMap = make(map[string]csModels.SnapshotPublic)
	}

	return cs.SnapshotMap
}

func (cs *VmCS) GetPortForwardingRule(name string) *csModels.PortForwardingRulePublic {
	resource, ok := cs.GetPortForwardingRuleMap()[name]
	if !ok {
		return nil
	}

	return &resource
}

func (cs *VmCS) GetSnapshotByID(id string) *csModels.SnapshotPublic {
	resource, ok := cs.GetSnapshotMap()[id]
	if !ok {
		return nil
	}

	return &resource
}

func (cs *VmCS) GetSnapshotByName(name string) *csModels.SnapshotPublic {
	for _, resource := range cs.GetSnapshotMap() {
		if resource.Name == name {
			return &resource
		}
	}

	return nil
}

func (cs *VmCS) SetSnapshot(name string, resource csModels.SnapshotPublic) {
	cs.GetSnapshotMap()[name] = resource
}

func (cs *VmCS) SetPortForwardingRule(name string, resource csModels.PortForwardingRulePublic) {
	cs.GetPortForwardingRuleMap()[name] = resource
}

func (cs *VmCS) DeleteSnapshot(name string) {
	delete(cs.GetSnapshotMap(), name)
}

func (cs *VmCS) DeletePortForwardingRule(name string) {
	delete(cs.GetPortForwardingRuleMap(), name)
}

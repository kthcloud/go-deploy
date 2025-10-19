package generators

import (
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/models"
)

// K8sGenerator is a generator for K8s resources
// It is used to generate the `publics`, such as models.DeploymentPublic and models.IngressPublic
type K8sGenerator interface {
	// Namespace returns a models.NamespacePublic that should be created
	Namespace() *models.NamespacePublic

	// Deployments returns a list of models.DeploymentPublic that should be created
	Deployments() []models.DeploymentPublic

	// VM returns a models.VmPublic that should be created
	VM() *models.VmPublic

	// Services returns a list of models.ServicePublic that should be created
	Services() []models.ServicePublic

	// Ingresses returns a list of models.IngressPublic that should be created
	Ingresses() []models.IngressPublic

	// PVs returns a list of models.PvPublic that should be created
	PVs() []models.PvPublic

	// PVCs returns a list of models.PvcPublic that should be created
	PVCs() []models.PvcPublic

	// RCTs returns a list of models.ResourceClaimTemplatePublic that should be created
	RCTs() []models.ResourceClaimTemplatePublic

	// Secrets returns a list of models.SecretPublic that should be created
	Secrets() []models.SecretPublic

	// OneShotJobs returns a list of models.JobPublic that should be created
	OneShotJobs() []models.JobPublic

	// HPAs returns a list of models.HpaPublic that should be created
	HPAs() []models.HpaPublic

	// NetworkPolicies returns a list of models.NetworkPolicyPublic that should be created
	NetworkPolicies() []models.NetworkPolicyPublic
}

type K8sGeneratorBase struct {
	K8sGenerator
}

func (kg *K8sGeneratorBase) Namespace() *models.NamespacePublic {
	return nil
}

func (kg *K8sGeneratorBase) Deployments() []models.DeploymentPublic {
	return nil
}

func (kg *K8sGeneratorBase) VMs() []models.VmPublic {
	return nil
}

func (kg *K8sGeneratorBase) Services() []models.ServicePublic {
	return nil
}

func (kg *K8sGeneratorBase) Ingresses() []models.IngressPublic {
	return nil
}

func (kg *K8sGeneratorBase) PVs() []models.PvPublic {
	return nil
}

func (kg *K8sGeneratorBase) PVCs() []models.PvcPublic {
	return nil
}

func (kg *K8sGeneratorBase) RCTs() []models.ResourceClaimTemplatePublic {
	return nil
}

func (kg *K8sGeneratorBase) Secrets() []models.SecretPublic {
	return nil
}

func (kg *K8sGeneratorBase) HPAs() []models.HpaPublic {
	return nil
}

func (kg *K8sGeneratorBase) NetworkPolicies() []models.NetworkPolicyPublic {
	return nil
}

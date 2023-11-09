package resources

import (
	"go-deploy/models/config"
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/models/sys/deployment/storage_manager"
	vmModels "go-deploy/models/sys/vm"
	"go-deploy/pkg/subsystems/k8s"
)

type Deployment struct {
	deployment *deploymentModels.Deployment
	zone       *config.DeploymentZone
}

type StorageManager struct {
	storageManager *storage_manager.StorageManager
	zone           *config.DeploymentZone
}

type VM struct {
	vm             *vmModels.VM
	vmZone         *config.VmZone
	deploymentZone *config.DeploymentZone
}

type PublicGeneratorType struct {
	d Deployment
	v VM
	s StorageManager
}

func PublicGenerator() *PublicGeneratorType {
	return &PublicGeneratorType{}
}

func (pc *PublicGeneratorType) WithDeploymentZone(zone *config.DeploymentZone) *PublicGeneratorType {
	pc.d.zone = zone
	pc.s.zone = zone
	pc.v.deploymentZone = zone
	return pc
}

func (pc *PublicGeneratorType) WithVmZone(zone *config.VmZone) *PublicGeneratorType {
	pc.v.vmZone = zone
	return pc
}

func (pc *PublicGeneratorType) WithDeployment(deployment *deploymentModels.Deployment) *PublicGeneratorType {
	pc.d.deployment = deployment
	return pc
}

func (pc *PublicGeneratorType) WithStorageManager(storageManager *storage_manager.StorageManager) *PublicGeneratorType {
	pc.s.storageManager = storageManager
	return pc
}

func (pc *PublicGeneratorType) WithVM(vm *vmModels.VM) *PublicGeneratorType {
	pc.v.vm = vm
	return pc
}

func (pc *PublicGeneratorType) K8s(client *k8s.Client) *K8sGenerator {
	return &K8sGenerator{
		PublicGeneratorType: pc,
		namespace:           client.Namespace,
		client:              client,
	}
}

func (pc *PublicGeneratorType) Harbor(project string) *HarborGenerator {
	return &HarborGenerator{
		PublicGeneratorType: pc,
		project:             project,
	}
}

func (pc *PublicGeneratorType) GitHub(token string, repositoryID int64) *GitHubGenerator {
	return &GitHubGenerator{
		PublicGeneratorType: pc,
		token:               token,
		repositoryID:        repositoryID,
	}
}

func (pc *PublicGeneratorType) CS() *CsGenerator {
	return &CsGenerator{
		PublicGeneratorType: pc,
	}
}

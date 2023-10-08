package resources

import (
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/models/sys/deployment/storage_manager"
	"go-deploy/models/sys/enviroment"
)

type PublicGeneratorType struct {
	zone *enviroment.DeploymentZone

	deployment     *deploymentModels.Deployment
	storageManager *storage_manager.StorageManager

	createParams *deploymentModels.CreateParams
	updateParams *deploymentModels.UpdateParams

	storageManagerCreateParams *storage_manager.CreateParams
}

func PublicGenerator() *PublicGeneratorType {
	return &PublicGeneratorType{}
}

func (pc *PublicGeneratorType) WithDeployment(deployment *deploymentModels.Deployment) *PublicGeneratorType {
	pc.deployment = deployment
	return pc
}

func (pc *PublicGeneratorType) WithStorageManager(storageManager *storage_manager.StorageManager) *PublicGeneratorType {
	pc.storageManager = storageManager
	return pc
}

func (pc *PublicGeneratorType) WithCreateParams(params *deploymentModels.CreateParams) *PublicGeneratorType {
	pc.createParams = params
	return pc
}

func (pc *PublicGeneratorType) WithUpdateParams(params *deploymentModels.UpdateParams) *PublicGeneratorType {
	pc.updateParams = params
	return pc
}

func (pc *PublicGeneratorType) WithStorageManagerCreateParams(params *storage_manager.CreateParams) *PublicGeneratorType {
	pc.storageManagerCreateParams = params
	return pc
}

func (pc *PublicGeneratorType) K8s(namespace string) *K8sGenerator {
	return &K8sGenerator{
		PublicGeneratorType: pc,
		namespace:           namespace,
	}
}

func (pc *PublicGeneratorType) Harbor(project string) *HarborGenerator {
	return &HarborGenerator{
		PublicGeneratorType: pc,
		project:             project,
	}
}

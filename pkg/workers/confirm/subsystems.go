package confirm

import (
	deploymentModels "go-deploy/models/sys/deployment"
	smModels "go-deploy/models/sys/sm"
	vmModels "go-deploy/models/sys/vm"
)

func getDeploymentDeletedConfirmers() []func(*deploymentModels.Deployment) (bool, error) {
	return []func(*deploymentModels.Deployment) (bool, error){
		k8sDeletedDeployment,
		harborDeleted,
		gitHubDeleted,
	}
}

func getSmDeletedConfirmers() []func(*smModels.SM) (bool, error) {
	return []func(*smModels.SM) (bool, error){
		k8sDeletedSM,
	}
}

func getVmDeletedConfirmers() []func(*vmModels.VM) (bool, error) {
	return []func(*vmModels.VM) (bool, error){
		csDeleted,
		gpuCleared,
		k8sDeletedVM,
	}
}

func DeploymentDeleted(deployment *deploymentModels.Deployment) bool {
	confirmers := getDeploymentDeletedConfirmers()
	for _, confirmer := range confirmers {
		deleted, _ := confirmer(deployment)
		if !deleted {
			return false
		}
	}
	return true
}

func SmDeleted(sm *smModels.SM) bool {
	confirmers := getSmDeletedConfirmers()
	for _, confirmer := range confirmers {
		deleted, _ := confirmer(sm)
		if !deleted {
			return false
		}
	}
	return true
}

func VmDeleted(vm *vmModels.VM) bool {
	confirmers := getVmDeletedConfirmers()
	for _, confirmer := range confirmers {
		deleted, _ := confirmer(vm)
		if !deleted {
			return false
		}
	}
	return true
}

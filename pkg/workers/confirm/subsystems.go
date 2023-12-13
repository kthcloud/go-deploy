package confirm

import (
	deploymentModels "go-deploy/models/sys/deployment"
	smModels "go-deploy/models/sys/sm"
	vmModels "go-deploy/models/sys/vm"
)

func getDeploymentCreatedConfirmers() []func(*deploymentModels.Deployment) (bool, error) {
	return []func(*deploymentModels.Deployment) (bool, error){
		k8sCreatedDeployment,
		harborCreated,
		gitHubCreated,
	}
}

func getDeploymentDeletedConfirmers() []func(*deploymentModels.Deployment) (bool, error) {
	return []func(*deploymentModels.Deployment) (bool, error){
		k8sDeletedDeployment,
		harborDeleted,
		gitHubDeleted,
	}
}

func getSmCreatedConfirmers() []func(*smModels.SM) (bool, error) {
	return []func(*smModels.SM) (bool, error){
		k8sCreatedSM,
	}
}

func getSmDeletedConfirmers() []func(*smModels.SM) (bool, error) {
	return []func(*smModels.SM) (bool, error){
		k8sDeletedSM,
	}
}

func getVmCreatedConfirmers() []func(*vmModels.VM) (bool, error) {
	return []func(*vmModels.VM) (bool, error){
		csCreated,
		k8sCreatedVM,
	}
}

func getVmDeletedConfirmers() []func(*vmModels.VM) (bool, error) {
	return []func(*vmModels.VM) (bool, error){
		csDeleted,
		gpuCleared,
		k8sDeletedVM,
	}
}

func DeploymentCreated(deployment *deploymentModels.Deployment) bool {
	confirmers := getDeploymentCreatedConfirmers()
	for _, confirmer := range confirmers {
		created, _ := confirmer(deployment)
		if !created {
			return false
		}
	}
	return true
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

func SmCreated(sm *smModels.SM) bool {
	confirmers := getSmCreatedConfirmers()
	for _, confirmer := range confirmers {
		created, _ := confirmer(sm)
		if !created {
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

func VmCreated(vm *vmModels.VM) bool {
	confirmers := getVmCreatedConfirmers()
	for _, confirmer := range confirmers {
		created, _ := confirmer(vm)
		if !created {
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

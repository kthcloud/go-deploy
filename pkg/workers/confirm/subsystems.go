package confirm

import (
	deploymentModels "go-deploy/models/sys/deployment"
	smModels "go-deploy/models/sys/sm"
	vmModels "go-deploy/models/sys/vm"
)

// getDeploymentDeletedConfirmers gets the confirmers for deployment deletion.
func getDeploymentDeletedConfirmers() []func(*deploymentModels.Deployment) (bool, error) {
	return []func(*deploymentModels.Deployment) (bool, error){
		k8sDeletedDeployment,
		harborDeleted,
		gitHubDeleted,
	}
}

// getSmDeletedConfirmers gets the confirmers for SM deletion.
func getSmDeletedConfirmers() []func(*smModels.SM) (bool, error) {
	return []func(*smModels.SM) (bool, error){
		k8sDeletedSM,
	}
}

// getVmDeletedConfirmers gets the confirmers for VM deletion.
func getVmDeletedConfirmers() []func(*vmModels.VM) (bool, error) {
	return []func(*vmModels.VM) (bool, error){
		csDeleted,
		gpuCleared,
		portsCleared,
		k8sDeletedVM,
	}
}

// DeploymentDeleted checks if the deployment is deleted by checking all confirmers.
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

// SmDeleted checks if the SM is deleted by checking all confirmers.
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

// VmDeleted checks if the VM is deleted by checking all confirmers.
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

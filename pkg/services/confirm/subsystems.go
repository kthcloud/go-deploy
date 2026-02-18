package confirm

import "github.com/kthcloud/go-deploy/models/model"

// getDeploymentDeletedConfirmers gets the confirmers for deployment deletion.
func getDeploymentDeletedConfirmers() []func(*model.Deployment) (bool, error) {
	return []func(*model.Deployment) (bool, error){
		k8sDeletedDeployment,
		harborDeleted,
	}
}

// getSmDeletedConfirmers gets the confirmers for SM deletion.
func getSmDeletedConfirmers() []func(*model.SM) (bool, error) {
	return []func(*model.SM) (bool, error){
		k8sDeletedSM,
	}
}

// getVmDeletedConfirmers gets the confirmers for VM deletion.
func getVmDeletedConfirmers() []func(*model.VM) (bool, error) {
	return []func(*model.VM) (bool, error){
		k8sDeletedVM,
		portsCleared,
	}
}

// getGCDeletedConfirmers gets the confirmers for GC deletion.
func getGCDeletedConfirmers() []func(*model.GpuClaim) (bool, error) {
	return []func(*model.GpuClaim) (bool, error){
		k8sDeletedGC,
	}
}

// DeploymentDeleted checks if the deployment is deleted by checking all confirmers.
func DeploymentDeleted(deployment *model.Deployment) bool {
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
func SmDeleted(sm *model.SM) bool {
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
func VmDeleted(vm *model.VM) bool {
	confirmers := getVmDeletedConfirmers()
	for _, confirmer := range confirmers {
		deleted, _ := confirmer(vm)
		if !deleted {
			return false
		}
	}
	return true
}

// GCDeleted checks if the GC is deleted by checking all confirmers.
func GCDeleted(gc *model.GpuClaim) bool {
	confirmers := getGCDeletedConfirmers()
	for _, confirmer := range confirmers {
		deleted, _ := confirmer(gc)
		if !deleted {
			return false
		}
	}
	return true
}

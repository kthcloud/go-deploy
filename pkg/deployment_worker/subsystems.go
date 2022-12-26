package deployment_worker

import (
	"go-deploy/models/deployment"
)

func getCreatedConfirmers() []func(*deployment.Deployment) (bool, error) {
	return []func(*deployment.Deployment) (bool, error){
		K8sCreated,
		NPMCreated,
		HarborCreated,
	}
}

func getDeletedConfirmers() []func(*deployment.Deployment) (bool, error) {
	return []func(*deployment.Deployment) (bool, error){
		K8sDeleted,
		NPMDeleted,
		HarborDeleted,
	}
}

func Created(deployment *deployment.Deployment) bool {
	confirmers := getCreatedConfirmers()
	for _, confirmer := range confirmers {
		created, _ := confirmer(deployment)
		if !created {
			return false
		}
	}
	return true
}

func Deleted(deployment *deployment.Deployment) bool {
	confirmers := getDeletedConfirmers()
	for _, confirmer := range confirmers {
		deleted, _ := confirmer(deployment)
		if !deleted {
			return false
		}
	}
	return true
}

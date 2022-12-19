package deployment_worker

import "go-deploy/models"

func getCreatedConfirmers() []func(*models.Deployment) (bool, error) {
	return []func(*models.Deployment) (bool, error){
		K8sCreated,
		NPMCreated,
		HarborCreated,
	}
}

func getDeletedConfirmers() []func(*models.Deployment) (bool, error) {
	return []func(*models.Deployment) (bool, error){
		K8sDeleted,
		NPMDeleted,
		HarborDeleted,
	}
}

func Created(deployment *models.Deployment) bool {
	confirmers := getCreatedConfirmers()
	for _, confirmer := range confirmers {
		created, _ := confirmer(deployment)
		if !created {
			return false
		}
	}
	return true
}

func Deleted(deployment *models.Deployment) bool {
	confirmers := getDeletedConfirmers()
	for _, confirmer := range confirmers {
		deleted, _ := confirmer(deployment)
		if !deleted {
			return false
		}
	}
	return true
}

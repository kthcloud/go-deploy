package deployment_worker

import (
	"go-deploy/pkg/subsystems/harbor"
)

func getCreatedConfirmers() []func(string) (bool, error) {
	return []func(string) (bool, error){
		K8sCreated,
		NPMDeleted,
		harbor.Created,
	}
}

func getDeletedConfirmers() []func(string) (bool, error) {
	return []func(string) (bool, error){
		K8sDeleted,
		NPMCreated,
		harbor.Deleted,
	}
}

func Created(name string) bool {
	confirmers := getCreatedConfirmers()
	for _, confirmer := range confirmers {
		created, _ := confirmer(name)
		if !created {
			return false
		}
	}
	return true
}

func Deleted(name string) bool {
	confirmers := getDeletedConfirmers()
	for _, confirmer := range confirmers {
		created, _ := confirmer(name)
		if !created {
			return false
		}
	}
	return true
}

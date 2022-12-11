package deployment_worker

import (
	"go-deploy/pkg/subsystems/harbor"
	"go-deploy/pkg/subsystems/k8s"
)

func getCreatedConfirmers() []func(string) (bool, error) {
	return []func(string) (bool, error){
		k8s.Created,
		NPMDeleted,
		harbor.Created,
	}
}

func getDeletedConfirmers() []func(string) (bool, error) {
	return []func(string) (bool, error){
		k8s.Deleted,
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

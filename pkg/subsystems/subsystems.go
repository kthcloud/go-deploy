package subsystems

import (
	"deploy-api-go/pkg/subsystems/harbor"
	"deploy-api-go/pkg/subsystems/k8s"
	"deploy-api-go/pkg/subsystems/npm"
)

func getCreatedConfirmers() []func(string) (bool, error) {
	return []func(string) (bool, error){
		k8s.Created,
		npm.Created,
		harbor.Created,
	}
}

func getDeletedConfirmers() []func(string) (bool, error) {
	return []func(string) (bool, error){
		k8s.Deleted,
		npm.Deleted,
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

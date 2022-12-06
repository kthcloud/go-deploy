package k8s

import (
	"fmt"
	"go-deploy/utils/subsystemutils"
)

func deletedNamespace(name string) (bool, error) {
	created, err := createdNamespace(name)
	return !created, err
}

func Deleted(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if k8s setup is deleted for project %s. details: %s", name, err)
	}

	namespaceDeleted, err := deletedNamespace(subsystemutils.GetPrefixedName(name))
	if err != nil {
		return false, makeError(err)
	}

	return namespaceDeleted, nil
}

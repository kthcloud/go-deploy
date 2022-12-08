package harbor

import (
	"context"
	"fmt"
	"go-deploy/utils/subsystemutils"
	"log"
	"strings"
)

func deleteProject(name string) error {
	prefixedName := subsystemutils.GetPrefixedName(name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete harbor project %s. details: %s", prefixedName, err)
	}

	client, err := createClient()
	if err != nil {
		return makeError(err)
	}

	err = client.DeleteProject(context.TODO(), prefixedName)
	if err != nil {
		errString := fmt.Sprintf("%s", err)
		if strings.Contains(errString, "deleteProjectPreconditionFailed") {
			return nil
		}

		return makeError(err)
	}

	return nil
}

func deleteRepository(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete harbor repository %s. details: %s", name, err)
	}

	client, err := createClient()
	if err != nil {
		return makeError(err)
	}

	exists, project, err := assertProjectExists(client, subsystemutils.GetPrefixedName(name))
	if !exists {
		return nil
	}

	err = client.DeleteRepository(context.TODO(), project.Name, name)
	if err != nil {
		errStr := fmt.Sprintf("%s", err)
		if !strings.Contains(errStr, "deleteRepositoryNotFound") {
			return makeError(err)
		}
	}

	return nil
}

func Delete(name string) error {
	log.Println("deleting harbor setup for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete harbor setup for deployment %s. details: %s", name, err)
	}

	err := deleteRepository(name)
	if err != nil {
		return makeError(err)
	}

	err = deleteProject(name)
	if err != nil {
		return makeError(err)
	}

	return nil
}

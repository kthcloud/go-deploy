package harbor

import (
	"context"
	"errors"
	"fmt"
	harborErrors "github.com/mittwald/goharbor-client/v5/apiv2/pkg/errors"
	"go-deploy/utils/subsystemutils"
)

func deletedProject(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if harbor project %s is created. details: %s", subsystemutils.GetPrefixedName(name), err)
	}

	client, err := createClient()
	if err != nil {
		return false, makeError(err)
	}

	_, err = client.GetProject(context.TODO(), subsystemutils.GetPrefixedName(name))
	if err != nil {
		targetErr := &harborErrors.ErrProjectNotFound{}
		if !errors.As(err, &targetErr) {
			return false, makeError(err)
		}
	}

	return true, nil
}

func Deleted(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if harbor setup is created for deployment %s. details: %s", name, err)
	}

	projectDeleted, err := deletedProject(name)
	if err != nil {
		return false, makeError(err)
	}

	return projectDeleted, nil
}

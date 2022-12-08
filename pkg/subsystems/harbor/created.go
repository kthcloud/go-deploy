package harbor

import (
	"context"
	"fmt"
	"go-deploy/utils/subsystemutils"
)

func createdProject(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if harbor project %s is created. details: %s", subsystemutils.GetPrefixedName(name), err)
	}

	client, err := createClient()
	if err != nil {
		return false, makeError(err)
	}

	project, err := client.GetProject(context.TODO(), subsystemutils.GetPrefixedName(name))
	if err != nil {
		return false, makeError(err)
	}

	publicProject := project.Metadata.Public == "true"

	return project.ProjectID != 0 && publicProject, nil
}

func createdRobot(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if harbor robot %s is created. details: %s", getRobotFullName(name), err)
	}

	client, err := createClient()
	if err != nil {
		return false, makeError(err)
	}

	robot, err := getRobotByNameV1(client, subsystemutils.GetPrefixedName(name), getRobotFullName(name))
	if err != nil {
		return false, makeError(err)
	}

	return robot.ID != 0, nil
}

func createdRepository(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if harbor repository %s is created. details: %s", name, err)
	}

	client, err := createClient()
	if err != nil {
		return false, makeError(err)
	}

	repository, err := client.GetRepository(context.TODO(), subsystemutils.GetPrefixedName(name), name)
	if err != nil {
		return false, makeError(err)
	}

	return repository.ID != 0, nil
}

func createdWebhook(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if harbor repository %s is created. details: %s", name, err)
	}

	client, err := createClient()
	if err != nil {
		return false, makeError(err)
	}

	project, err := client.GetProject(context.TODO(), subsystemutils.GetPrefixedName(name))
	if err != nil {
		return false, makeError(err)
	}

	if project == nil {
		return false, nil
	}

	webhookPolicies, err := client.ListProjectWebhookPolicies(context.TODO(), int(project.ProjectID))
	if err != nil {
		return false, makeError(err)
	}

	for _, policy := range webhookPolicies {
		if policy.Name == name {
			return true, nil
		}
	}
	return false, nil
}

func Created(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if harbor setup is created for deployment %s. details: %s", name, err)
	}

	projectCreated, err := createdProject(name)
	if err != nil {
		return false, makeError(err)
	}

	if !projectCreated {
		return false, nil
	}

	robotCreated, err := createdRobot(name)
	if err != nil {
		return false, makeError(err)
	}

	repositoryCreated, err := createdRepository(name)
	if err != nil {
		return false, makeError(err)
	}

	webhookCreated, err := createdWebhook(name)
	if err != nil {
		return false, makeError(err)
	}

	return robotCreated && repositoryCreated && webhookCreated, nil
}

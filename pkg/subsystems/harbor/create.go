package harbor

import (
	"context"
	"errors"
	"fmt"
	modelv2 "github.com/mittwald/goharbor-client/v5/apiv2/model"
	"github.com/mittwald/goharbor-client/v5/apiv2/pkg/clients/artifact"
	harborErrors "github.com/mittwald/goharbor-client/v5/apiv2/pkg/errors"
	"go-deploy/pkg/conf"
	"go-deploy/utils"
	"go-deploy/utils/subsystemutils"
	"log"
	"strings"
)

func createProject(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create harbor project %s. details: %s", name, err.Error())
	}

	client, err := createClient()
	if err != nil {
		return makeError(err)
	}

	prefixedName := subsystemutils.GetPrefixedName(name)
	project, err := client.GetProject(context.TODO(), prefixedName)
	if err != nil {
		targetErr := &harborErrors.ErrProjectNotFound{}
		if !errors.As(err, &targetErr) {
			return makeError(err)
		}
	}

	if project == nil {
		requestBody := createProjectRequestBody(subsystemutils.GetPrefixedName(name))
		err = client.NewProject(context.TODO(), &requestBody)
		if err != nil {
			return makeError(err)
		}

		err = client.UpdateProjectMetadata(context.TODO(), prefixedName, "public", "true")
		if err != nil {
			return makeError(err)
		}
	} else if project.Metadata.Public == "false" {
		err = client.UpdateProjectMetadata(context.TODO(), prefixedName, "public", "true")
		if err != nil {
			return makeError(err)
		}
	}
	return nil
}

func createRobot(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create harbor robot %s. details: %s", name, err)
	}

	client, err := createClient()
	if err != nil {
		return makeError(err)
	}

	prefixedName := subsystemutils.GetPrefixedName(name)
	projectExists, project, err := assertProjectExists(client, prefixedName)
	if err != nil {
		return makeError(err)
	}

	if !projectExists {
		err = fmt.Errorf("no project exists")
		return makeError(err)
	}

	robots, err := client.ListProjectRobotsV1(context.TODO(), project.Name)
	if err != nil {
		return err
	}

	robotResult := &modelv2.Robot{}
	for _, robot := range robots {
		if robot.Name == getRobotFullName(name) {
			robotResult = robot
			break
		}
	}

	if robotResult.ID != 0 {
		return nil
	}

	robotCreatedBody, err := createHarborRobot(name)
	if err != nil {
		return makeError(err)
	}

	err = updateDatabaseRobot(name, robotCreatedBody.Name, robotCreatedBody.Secret)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func createRepository(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create harbor repository for deployment %s. details: %s", name, err)
	}

	prefixedName := subsystemutils.GetPrefixedName(name)

	client, err := createClient()
	if err != nil {
		return makeError(err)
	}

	projectArtifact, err := client.GetArtifact(context.TODO(), prefixedName, name, "latest")
	if err != nil {
		errStr := fmt.Sprintf("%s", err)
		is404 := strings.Contains(errStr, "getArtifactNotFound")
		if !is404 {
			return makeError(err)
		}
	}

	if projectArtifact != nil {
		return nil
	}

	project, repository := subsystemutils.GetPlaceholderImage()

	placeholderArtifact, err := client.GetArtifact(context.TODO(), project, repository, "latest")
	if err != nil {
		return makeError(err)
	}

	copyRef := &artifact.CopyReference{
		ProjectName:    project,
		RepositoryName: repository,
		Tag:            "latest",
		Digest:         placeholderArtifact.Digest,
	}

	err = client.CopyArtifact(context.TODO(), copyRef, prefixedName, name)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func createWebhook(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create harbor webhook for %s. details: %s", name, err)
	}

	client, err := createClient()
	if err != nil {
		return makeError(err)
	}

	prefixedName := subsystemutils.GetPrefixedName(name)
	projectExists, project, err := assertProjectExists(client, prefixedName)
	if err != nil {
		return makeError(err)
	}

	if !projectExists {
		err = fmt.Errorf("no project exists")
		return makeError(err)
	}

	webhooks, err := client.ListProjectWebhookPolicies(context.TODO(), int(project.ProjectID))
	if err != nil {
		return makeError(err)
	}

	for _, hook := range webhooks {
		if hook.Name == name {
			return nil
		}
	}

	webhookToken, err := generateToken(conf.Env.Harbor.WebhookSecret)
	if err != nil {
		return makeError(err)
	}

	err = updateDatabaseWebhook(name, utils.HashString(webhookToken))
	if err != nil {
		return makeError(err)
	}

	webhookTargetAddress := fmt.Sprintf("%s/v1/hooks/deployments/harbor", conf.Env.ExternalUrl)

	err = client.AddProjectWebhookPolicy(context.TODO(), int(project.ProjectID), &modelv2.WebhookPolicy{
		Enabled:    true,
		EventTypes: getWebhookEventTypes(),
		Name:       name,
		Targets: []*modelv2.WebhookTargetObject{
			{
				Address:        webhookTargetAddress,
				AuthHeader:     createAuthHeader(webhookToken),
				SkipCertVerify: false,
				Type:           "http",
			},
		},
	})

	if err != nil {
		return makeError(err)
	}

	return nil
}

func Create(name string) error {
	log.Println("creating harbor setup for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to create harbor setup for deployment %s. details: %s", name, err)
	}

	err := createProject(name)
	if err != nil {
		return makeError(err)
	}

	err = createRobot(name)
	if err != nil {
		return makeError(err)
	}

	err = createRepository(name)
	if err != nil {
		return makeError(err)
	}
	err = createWebhook(name)
	if err != nil {
		return makeError(err)
	}

	return nil
}

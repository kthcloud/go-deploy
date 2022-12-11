package subsystem_service

import (
	"fmt"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/harbor"
	"go-deploy/pkg/subsystems/harbor/models"
	"go-deploy/utils/subsystemutils"
	"log"
)

func CreateHarbor(name string) error {
	log.Println("creating harbor setup for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup harbor for deployment %s. details: %s", name, err)
	}

	client, err := harbor.New(&harbor.ClientConf{
		ApiUrl:        conf.Env.Harbor.Url,
		Username:      conf.Env.Harbor.Identity,
		Password:      conf.Env.Harbor.Secret,
		WebhookSecret: conf.Env.Harbor.WebhookSecret,
	})
	if err != nil {
		return makeError(err)
	}

	projectName := subsystemutils.GetPrefixedName(name)

	err = client.CreateProject(projectName)
	if err != nil {
		return makeError(err)
	}

	err = client.CreateRobot(projectName, name)
	if err != nil {
		return makeError(err)
	}

	err = client.CreateRepository(projectName, name, &models.PlaceHolder{
		ProjectName: conf.Env.DockerRegistry.PlaceHolderProject,
		Repository:  conf.Env.DockerRegistry.PlaceHolderRepository,
	})
	if err != nil {
		return makeError(err)
	}

	webhookTarget := fmt.Sprintf("%s/v1/hooks/deployments/harbor", conf.Env.ExternalUrl)
	err = client.CreateWebhook(projectName, name, webhookTarget)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func DeleteHarbor(name string) error {
	log.Println("deleting harbor setup for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete harbor setup for deployment %s. details: %s", name, err)
	}

	client, err := harbor.New(&harbor.ClientConf{
		ApiUrl:        conf.Env.Harbor.Url,
		Username:      conf.Env.Harbor.Identity,
		Password:      conf.Env.Harbor.Secret,
		WebhookSecret: conf.Env.Harbor.WebhookSecret,
	})
	if err != nil {
		return makeError(err)
	}

	projectName := subsystemutils.GetPrefixedName(name)

	err = client.DeleteRepository(projectName, name)
	if err != nil {
		return makeError(err)
	}

	err = client.DeleteProject(projectName)
	if err != nil {
		return makeError(err)
	}

	return nil
}

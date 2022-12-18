package deployment_worker

import (
	"fmt"
	"go-deploy/models"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/harbor"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/pkg/subsystems/npm"
	"go-deploy/utils/subsystemutils"
)

func getFQDN(name string) string {
	return fmt.Sprintf("%s.%s", name, conf.Env.ParentDomain)
}

func NPMCreated(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if npm setup is created for deployment %s. details: %s", name, err)
	}

	client, err := npm.New(&npm.ClientConf{
		ApiUrl:   conf.Env.NPM.Url,
		Username: conf.Env.NPM.Identity,
		Password: conf.Env.NPM.Secret,
	})
	if err != nil {
		return false, makeError(err)
	}

	deployment, err := models.GetDeploymentByName(name)

	if deployment.Subsytems.Npm.Public.ID == 0 {
		return false, nil
	}

	proxyHostCreated, err := client.ProxyHostCreated(deployment.Subsytems.Npm.Public.ID)
	if err != nil {
		return false, makeError(err)
	}

	return proxyHostCreated, nil
}

func NPMDeleted(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if npm setup is deleted for deployment %s. details: %s", name, err)
	}

	deployment, err := models.GetDeploymentByName(name)

	if deployment.Subsytems.Npm.Public.ID == 0 {
		return true, nil
	}

	client, err := npm.New(&npm.ClientConf{
		ApiUrl:   conf.Env.NPM.Url,
		Username: conf.Env.NPM.Identity,
		Password: conf.Env.NPM.Secret,
	})
	if err != nil {
		return false, makeError(err)
	}

	proxyHostDeleted, err := client.ProxyHostDeleted(deployment.Subsytems.Npm.Public.ID)
	if err != nil {
		return false, makeError(err)
	}

	return proxyHostDeleted, nil
}

func K8sCreated(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if k8s setup is created for deployment %s. details: %s", name, err)
	}

	client, err := k8s.New(&k8s.ClientConf{K8sAuth: conf.Env.K8s.Config})
	if err != nil {
		return false, makeError(err)
	}

	namespace := subsystemutils.GetPrefixedName(name)

	namespaceCreated, err := client.NamespaceCreated(namespace)
	if err != nil {
		return false, makeError(err)
	}

	if !namespaceCreated {
		return false, nil
	}

	deploymentCreated, err := client.DeploymentCreated(namespace, name)
	if err != nil {
		return false, makeError(err)
	}

	serviceCreated, err := client.ServiceCreated(namespace, name)
	if err != nil {
		return false, makeError(err)
	}

	return deploymentCreated && serviceCreated, nil
}

func K8sDeleted(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if k8s setup is deleted for deployment %s. details: %s", name, err)
	}

	client, err := k8s.New(&k8s.ClientConf{K8sAuth: conf.Env.K8s.Config})
	if err != nil {
		return false, makeError(err)
	}

	namespaceDeleted, err := client.NamespaceDeleted(subsystemutils.GetPrefixedName(name))
	if err != nil {
		return false, makeError(err)
	}

	return namespaceDeleted, nil
}

func HarborCreated(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if harbor setup is created for deployment %s. details: %s", name, err)
	}

	client, err := harbor.New(&harbor.ClientConf{
		ApiUrl:        conf.Env.Harbor.Url,
		Username:      conf.Env.Harbor.Identity,
		Password:      conf.Env.Harbor.Secret,
		WebhookSecret: conf.Env.Harbor.WebhookSecret,
	})
	if err != nil {
		return false, makeError(err)
	}

	projectName := subsystemutils.GetPrefixedName(name)

	projectCreated, err := client.ProjectCreated(projectName)
	if err != nil {
		return false, makeError(err)
	}

	if !projectCreated {
		return false, nil
	}

	robotCreated, err := client.RobotCreated(projectName, name)
	if err != nil {
		return false, makeError(err)
	}

	repositoryCreated, err := client.RepositoryCreated(projectName, name)
	if err != nil {
		return false, makeError(err)
	}

	webhookCreated, err := client.WebhookCreated(projectName, name)
	if err != nil {
		return false, makeError(err)
	}

	return robotCreated && repositoryCreated && webhookCreated, nil
}

func HarborDeleted(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if harbor setup is created for deployment %s. details: %s", name, err)
	}

	client, err := harbor.New(&harbor.ClientConf{
		ApiUrl:        conf.Env.Harbor.Url,
		Username:      conf.Env.Harbor.Identity,
		Password:      conf.Env.Harbor.Secret,
		WebhookSecret: conf.Env.Harbor.WebhookSecret,
	})
	if err != nil {
		return false, makeError(err)
	}

	projectDeleted, err := client.ProjectDeleted(name)
	if err != nil {
		return false, makeError(err)
	}

	return projectDeleted, nil
}

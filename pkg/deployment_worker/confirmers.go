package deployment_worker

import (
	"fmt"
	"go-deploy/models/deployment"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/utils/subsystemutils"
)

func getFQDN(name string) string {
	return fmt.Sprintf("%s.%s", name, conf.Env.ParentDomain)
}

func NPMCreated(deployment *deployment.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if npm setup is created for v1_deployment %s. details: %s", deployment.Name, err)
	}

	return deployment.Subsystems.Npm.ProxyHost.ID != 0, nil
}

func NPMDeleted(deployment *deployment.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if npm setup is deleted for v1_deployment %s. details: %s", deployment.Name, err)
	}

	return deployment.Subsystems.Npm.ProxyHost.ID == 0, nil
}

func K8sCreated(deployment *deployment.Deployment) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if k8s setup is created for v1_deployment %s. details: %s", deployment.Name, err)
	}

	client, err := k8s.New(&k8s.ClientConf{K8sAuth: conf.Env.K8s.Config})
	if err != nil {
		return false, makeError(err)
	}

	namespace := subsystemutils.GetPrefixedName(deployment.Name)

	namespaceCreated, err := client.NamespaceCreated(namespace)
	if err != nil {
		return false, makeError(err)
	}

	if !namespaceCreated {
		return false, nil
	}

	deploymentCreated, err := client.DeploymentCreated(namespace, deployment.Name)
	if err != nil {
		return false, makeError(err)
	}

	serviceCreated, err := client.ServiceCreated(namespace, deployment.Name)
	if err != nil {
		return false, makeError(err)
	}

	return deploymentCreated && serviceCreated, nil
}

func K8sDeleted(deployment *deployment.Deployment) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if k8s setup is deleted for v1_deployment %s. details: %s", deployment.Name, err)
	}

	client, err := k8s.New(&k8s.ClientConf{K8sAuth: conf.Env.K8s.Config})
	if err != nil {
		return false, makeError(err)
	}

	namespaceDeleted, err := client.NamespaceDeleted(subsystemutils.GetPrefixedName(deployment.Name))
	if err != nil {
		return false, makeError(err)
	}

	return namespaceDeleted, nil
}

func HarborCreated(deployment *deployment.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if harbor setup is created for v1_deployment %s. details: %s", deployment.Name, err)
	}

	harbor := &deployment.Subsystems.Harbor
	return harbor.Project.ID != 0 &&
		harbor.Robot.ID != 0 &&
		harbor.Repository.ID != 0 &&
		harbor.Webhook.ID != 0, nil
}

func HarborDeleted(deployment *deployment.Deployment) (bool, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to check if harbor setup is created for v1_deployment %s. details: %s", deployment.Name, err)
	}

	harbor := &deployment.Subsystems.Harbor
	return harbor.Project.ID == 0 &&
		harbor.Robot.ID == 0 &&
		harbor.Repository.ID == 0 &&
		harbor.Webhook.ID == 0, nil
}

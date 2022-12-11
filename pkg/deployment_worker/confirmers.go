package deployment_worker

import (
	"fmt"
	"go-deploy/pkg/conf"
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
		ApiUrl:   conf.Env.Npm.Url,
		Username: conf.Env.Npm.Identity,
		Password: conf.Env.Npm.Secret,
	})
	if err != nil {
		return false, makeError(err)
	}

	proxyHostCreated, err := client.ProxyHostCreated(getFQDN(name))
	if err != nil {
		return false, makeError(err)
	}

	return proxyHostCreated, nil
}

func NPMDeleted(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if npm setup is deleted for deployment %s. details: %s", name, err)
	}

	client, err := npm.New(&npm.ClientConf{
		ApiUrl:   conf.Env.Npm.Url,
		Username: conf.Env.Npm.Identity,
		Password: conf.Env.Npm.Secret,
	})
	if err != nil {
		return false, makeError(err)
	}

	proxyHostDeleted, err := client.ProxyHostDeleted(getFQDN(name))
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

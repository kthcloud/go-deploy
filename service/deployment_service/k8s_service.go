package deployment_service

import (
	"fmt"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/utils/subsystemutils"
	apiv1 "k8s.io/api/core/v1"
	"log"
	"strconv"
)

func CreateK8s(name string) error {
	log.Println("setting up k8s for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup k8s for deployment %s. details: %s", name, err)
	}

	client, err := k8s.New(&k8s.ClientConf{K8sAuth: conf.Env.K8s.Config})
	if err != nil {
		return makeError(err)
	}

	prefixedName := subsystemutils.GetPrefixedName(name)
	namespace := prefixedName
	dockerImage := fmt.Sprintf("%s/%s/%s", conf.Env.DockerRegistry.Url, prefixedName, name)
	port := conf.Env.AppPort

	err = client.CreateNamespace(namespace)
	if err != nil {
		return makeError(err)
	}

	err = client.CreateDeployment(namespace, name, dockerImage, []apiv1.EnvVar{
		{
			Name:  "DEPLOY_APP_PORT",
			Value: strconv.Itoa(port),
		},
	})
	if err != nil {
		return makeError(err)
	}

	err = client.CreateService(namespace, name, port, port)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func DeleteK8s(name string) error {
	log.Println("deleting k8s for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s for deployment %s. details: %s", name, err)
	}

	client, err := k8s.New(&k8s.ClientConf{K8sAuth: conf.Env.K8s.Config})
	if err != nil {
		return makeError(err)
	}

	err = client.DeleteNamespace(subsystemutils.GetPrefixedName(name))
	if err != nil {
		return makeError(err)
	}

	return nil
}

func RestartK8s(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to restart k8s %s. details: %s", name, err)
	}

	client, err := k8s.New(&k8s.ClientConf{K8sAuth: conf.Env.K8s.Config})
	if err != nil {
		return makeError(err)
	}

	err = client.RestartDeployment(subsystemutils.GetPrefixedName(name), name)
	if err != nil {
		return makeError(err)
	}

	return nil
}

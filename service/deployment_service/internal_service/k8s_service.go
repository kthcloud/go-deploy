package internal_service

import (
	"fmt"
	deploymentModel "go-deploy/models/deployment"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/k8s"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/utils/subsystemutils"
	"log"
	"strconv"
)

type K8sResult struct {
	Namespace  *k8sModels.NamespacePublic
	Deployment *k8sModels.DeploymentPublic
	Service    *k8sModels.ServicePublic
}

func createNamespacePublic(name string) *k8sModels.NamespacePublic {
	return &k8sModels.NamespacePublic{
		Name: name,
	}
}

func createDeploymentPublic(namespace, name, dockerImage string, envVars []k8sModels.EnvVar) *k8sModels.DeploymentPublic {
	return &k8sModels.DeploymentPublic{
		ID:          "",
		Name:        name,
		Namespace:   namespace,
		DockerImage: dockerImage,
		EnvVars:     envVars,
	}
}

func createServicePublic(namespace, name string, port, targetPort int) *k8sModels.ServicePublic {
	return &k8sModels.ServicePublic{
		ID:         "",
		Name:       name,
		Namespace:  namespace,
		Port:       port,
		TargetPort: targetPort,
	}
}

func CreateK8s(name string) (*K8sResult, error) {
	log.Println("setting up k8s for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup k8s for deployment %s. details: %s", name, err)
	}

	client, err := k8s.New(&k8s.ClientConf{
		K8sAuth: conf.Env.K8s.Config,
	})
	if err != nil {
		return nil, makeError(err)
	}

	port := conf.Env.AppPort

	deployment, err := deploymentModel.GetDeploymentByName(name)
	if err != nil {
		return nil, makeError(err)
	}

	// Namespace
	var namespace *k8sModels.NamespacePublic
	if deployment.Subsystems.K8s.Namespace.Name == "" {
		prefixedName := subsystemutils.GetPrefixedName(name)
		err = client.CreateNamespace(createNamespacePublic(prefixedName))
		if err != nil {
			return nil, makeError(err)
		}

		namespace = &k8sModels.NamespacePublic{}
		namespace.Name = prefixedName

		err = deploymentModel.UpdateSubsystemByName(name, "k8s", "namespace", namespace)
		if err != nil {
			return nil, makeError(err)
		}
	} else {
		namespace = &deployment.Subsystems.K8s.Namespace
	}

	// Deployment
	var k8sDeployment *k8sModels.DeploymentPublic
	if deployment.Subsystems.K8s.Deployment.ID == "" {
		prefixedName := subsystemutils.GetPrefixedName(name)
		dockerImage := fmt.Sprintf("%s/%s/%s", conf.Env.DockerRegistry.Url, prefixedName, name)

		id, err := client.CreateDeployment(createDeploymentPublic(namespace.Name, name, dockerImage, []k8sModels.EnvVar{
			{Name: "DEPLOY_APP_PORT", Value: strconv.Itoa(port)},
		}))
		if err != nil {
			return nil, makeError(err)
		}

		k8sDeployment, err = client.ReadDeployment(namespace.Name, id)
		if err != nil {
			return nil, makeError(err)
		}

		err = deploymentModel.UpdateSubsystemByName(name, "k8s", "deployment", k8sDeployment)
		if err != nil {
			return nil, makeError(err)
		}
	} else {
		k8sDeployment = &deployment.Subsystems.K8s.Deployment
	}

	// Service
	var service *k8sModels.ServicePublic
	if deployment.Subsystems.K8s.Service.ID == "" {
		id, err := client.CreateService(createServicePublic(namespace.Name, name, port, port))
		if err != nil {
			return nil, makeError(err)
		}

		service, err = client.ReadService(namespace.Name, id)
		if err != nil {
			return nil, makeError(err)
		}

		err = deploymentModel.UpdateSubsystemByName(name, "k8s", "service", service)
		if err != nil {
			return nil, makeError(err)
		}
	} else {
		service = &deployment.Subsystems.K8s.Service
	}

	return &K8sResult{
		Namespace:  namespace,
		Deployment: k8sDeployment,
		Service:    service,
	}, nil
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

	err = deploymentModel.UpdateSubsystemByName(name, "k8s", "service", k8sModels.ServicePublic{})
	if err != nil {
		return makeError(err)
	}

	err = deploymentModel.UpdateSubsystemByName(name, "k8s", "deployment", k8sModels.DeploymentPublic{})
	if err != nil {
		return makeError(err)
	}

	err = deploymentModel.UpdateSubsystemByName(name, "k8s", "namespace", k8sModels.NamespacePublic{})
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

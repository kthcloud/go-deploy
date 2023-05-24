package internal_service

import (
	"errors"
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
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

func createNamespacePublic(userID string) *k8sModels.NamespacePublic {
	return &k8sModels.NamespacePublic{
		ID:       "",
		Name:     userID,
		FullName: "",
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

func createIngressPublic(namespace, name, host, serviceName string, servicePort int) *k8sModels.IngressPublic {
	return &k8sModels.IngressPublic{
		ID:               "",
		Name:             name,
		Namespace:        namespace,
		ServiceName:      serviceName,
		ServicePort:      servicePort,
		IngressClassName: "caddy",
		Host:             host,
	}
}

func getExternalFQDN(name string) string {
	return fmt.Sprintf("%s.%s", name, conf.Env.Deployment.ParentDomain)
}

func CreateK8s(name, userID string, envs []deploymentModel.Env) (*K8sResult, error) {
	log.Println("setting up k8s for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup k8s for deployment %s. details: %s", name, err)
	}

	client, err := k8s.New(conf.Env.K8s.Client)
	if err != nil {
		return nil, makeError(err)
	}

	port := conf.Env.Deployment.Port

	deployment, err := deploymentModel.GetByName(name)
	if err != nil {
		return nil, makeError(err)
	}

	if deployment == nil {
		return nil, nil
	}

	// Namespace
	var namespace *k8sModels.NamespacePublic
	if deployment.Subsystems.K8s.Namespace.ID == "" {
		id, err := client.CreateNamespace(createNamespacePublic(userID))
		if err != nil {
			return nil, makeError(err)
		}

		namespace, err = client.ReadNamespace(id)
		if err != nil {
			return nil, makeError(err)
		}

		if namespace == nil {
			return nil, errors.New("failed to read namespace after creation")
		}

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
		dockerRegistryProject := subsystemutils.GetPrefixedName(userID)
		dockerImage := fmt.Sprintf("%s/%s/%s", conf.Env.DockerRegistry.URL, dockerRegistryProject, name)

		k8sEnvs := []k8sModels.EnvVar{
			{Name: "DEPLOY_APP_PORT", Value: strconv.Itoa(port)},
		}

		for _, env := range envs {
			k8sEnvs = append(k8sEnvs, k8sModels.EnvVar{
				Name:  env.Name,
				Value: env.Value,
			})
		}

		id, err := client.CreateDeployment(createDeploymentPublic(namespace.FullName, name, dockerImage, k8sEnvs))
		if err != nil {
			return nil, makeError(err)
		}

		k8sDeployment, err = client.ReadDeployment(namespace.FullName, id)
		if err != nil {
			return nil, makeError(err)
		}

		if k8sDeployment == nil {
			return nil, errors.New("failed to read deployment after creation")
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
		id, err := client.CreateService(createServicePublic(namespace.FullName, name, port, port))
		if err != nil {
			return nil, makeError(err)
		}

		service, err = client.ReadService(namespace.FullName, id)
		if err != nil {
			return nil, makeError(err)
		}

		if service == nil {
			return nil, errors.New("failed to read service after creation")
		}

		err = deploymentModel.UpdateSubsystemByName(name, "k8s", "service", service)
		if err != nil {
			return nil, makeError(err)
		}
	} else {
		service = &deployment.Subsystems.K8s.Service
	}

	// Ingress
	err = createIngress(client, deployment, namespace.FullName, service.Name, service.Port)
	if err != nil {
		return nil, makeError(err)
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

	client, err := k8s.New(conf.Env.K8s.Client)
	if err != nil {
		return makeError(err)
	}

	// delete everything in the opposite order of creation
	deployment, err := deploymentModel.GetByName(name)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		return nil
	}

	if deployment.Subsystems.K8s.Ingress.ID != "" {
		err = client.DeleteIngress(deployment.Subsystems.K8s.Ingress.Namespace, deployment.Subsystems.K8s.Ingress.ID)
		if err != nil {
			return makeError(err)
		}

		err = deploymentModel.UpdateSubsystemByName(name, "k8s", "ingress", k8sModels.IngressPublic{})
		if err != nil {
			return makeError(err)
		}
	}

	if deployment.Subsystems.K8s.Service.ID != "" {
		err = client.DeleteService(deployment.Subsystems.K8s.Service.Namespace, deployment.Subsystems.K8s.Service.ID)
		if err != nil {
			return makeError(err)
		}

		err = deploymentModel.UpdateSubsystemByName(name, "k8s", "service", k8sModels.ServicePublic{})
		if err != nil {
			return makeError(err)
		}
	}

	if deployment.Subsystems.K8s.Deployment.ID != "" {
		err = client.DeleteDeployment(deployment.Subsystems.K8s.Deployment.Namespace, deployment.Subsystems.K8s.Deployment.ID)
		if err != nil {
			return makeError(err)
		}

		err = deploymentModel.UpdateSubsystemByName(name, "k8s", "deployment", k8sModels.DeploymentPublic{})
		if err != nil {
			return makeError(err)
		}
	}

	if deployment.Subsystems.K8s.Namespace.ID != "" {
		err = deploymentModel.UpdateSubsystemByName(name, "k8s", "namespace", k8sModels.NamespacePublic{})
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func UpdateK8s(name string, params *deploymentModel.UpdateParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update k8s for deployment %s. details: %s", name, err)
	}

	if params == nil || (params.Envs == nil && params.Private == nil) {
		return nil
	}

	deployment, err := deploymentModel.GetByName(name)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		return nil
	}

	client, err := k8s.New(conf.Env.K8s.Client)
	if err != nil {
		return makeError(err)
	}

	if params.Envs != nil {
		if deployment.Subsystems.K8s.Deployment.ID != "" {
			k8sEnvs := []k8sModels.EnvVar{
				{Name: "DEPLOY_APP_PORT", Value: strconv.Itoa(conf.Env.Deployment.Port)},
			}
			for _, env := range *params.Envs {
				k8sEnvs = append(k8sEnvs, k8sModels.EnvVar{
					Name:  env.Name,
					Value: env.Value,
				})
			}

			deployment.Subsystems.K8s.Deployment.EnvVars = k8sEnvs

			err = client.UpdateDeployment(&deployment.Subsystems.K8s.Deployment)
			if err != nil {
				return makeError(err)
			}

			err = deploymentModel.UpdateSubsystemByName(name, "k8s", "deployment", &deployment.Subsystems.K8s.Deployment)
			if err != nil {
				return makeError(err)
			}
		}
	}

	if params.Private != nil {
		emptyOrPlaceHolder := isPlaceholderOrEmpty(deployment.Subsystems.K8s.Ingress.ID)

		if *params.Private && !emptyOrPlaceHolder ||
			(!*params.Private && emptyOrPlaceHolder) {
			if !emptyOrPlaceHolder {
				err = client.DeleteIngress(deployment.Subsystems.K8s.Ingress.Namespace, deployment.Subsystems.K8s.Ingress.ID)
				if err != nil {
					return makeError(err)
				}
			}

			err = deploymentModel.UpdateSubsystemByName(name, "k8s", "ingress", k8sModels.IngressPublic{})
			if err != nil {
				return makeError(err)
			}

			deployment, err = deploymentModel.GetByName(name)
			if err != nil {
				return makeError(err)
			}

			if *params.Private {
				err = deploymentModel.UpdateSubsystemByName(name, "k8s", "ingress", k8sModels.IngressPublic{
					ID: "placeholder",
				})
				if err != nil {
					return makeError(err)
				}
			} else {
				namespace := deployment.Subsystems.K8s.Namespace
				if namespace.ID == "" {
					return nil
				}

				service := deployment.Subsystems.K8s.Service
				if service.ID == "" {
					return nil
				}

				err = createIngress(client, deployment, deployment.Subsystems.K8s.Namespace.FullName, service.Name, service.Port)
				if err != nil {
					return makeError(err)
				}

			}
		}

	}

	return nil
}

func RestartK8s(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to restart k8s %s. details: %s", name, err)
	}

	deployment, err := deploymentModel.GetByName(name)
	if deployment.Subsystems.K8s.Deployment.ID == "" {
		return makeError(errors.New("can't restart deployment that is not yet created"))
	}

	client, err := k8s.New(conf.Env.K8s.Client)
	if err != nil {
		return makeError(err)
	}

	err = client.RestartDeployment(&deployment.Subsystems.K8s.Deployment)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func createIngress(client *k8s.Client, deployment *deploymentModel.Deployment, namespace, serviceName string, servicePort int) error {
	var ingress *k8sModels.IngressPublic
	if deployment.Subsystems.K8s.Ingress.ID == "" {
		id, err := client.CreateIngress(createIngressPublic(namespace, deployment.Name, getExternalFQDN(deployment.Name), serviceName, servicePort))
		if err != nil {
			return err
		}

		ingress, err = client.ReadIngress(namespace, id)
		if err != nil {
			return err
		}

		if ingress == nil {
			return errors.New("failed to read ingress after creation")
		}

		err = deploymentModel.UpdateSubsystemByName(deployment.Name, "k8s", "ingress", ingress)
		if err != nil {
			return err
		}
	}

	return nil
}

func isPlaceholderOrEmpty(id string) bool {
	return id == "" || id == "placeholder"
}

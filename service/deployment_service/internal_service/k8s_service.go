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
	"reflect"
	"strconv"
)

type K8sResult struct {
	Namespace  *k8sModels.NamespacePublic
	Deployment *k8sModels.DeploymentPublic
	Service    *k8sModels.ServicePublic
	Ingress    *k8sModels.IngressPublic
}

func createNamespacePublic(userID string) *k8sModels.NamespacePublic {
	return &k8sModels.NamespacePublic{
		ID:       "",
		Name:     userID,
		FullName: "",
	}
}

func createDeploymentPublic(namespace, name, dockerImage string, envs []deploymentModel.Env) *k8sModels.DeploymentPublic {
	port := conf.Env.Deployment.Port

	k8sEnvs := []k8sModels.EnvVar{
		{Name: "DEPLOY_APP_PORT", Value: strconv.Itoa(port)},
	}

	for _, env := range envs {
		k8sEnvs = append(k8sEnvs, k8sModels.EnvVar{
			Name:  env.Name,
			Value: env.Value,
		})
	}

	defaultLimits := k8sModels.Limits{
		CPU:    conf.Env.Deployment.Resources.Limits.CPU,
		Memory: conf.Env.Deployment.Resources.Limits.Memory,
	}

	defaultRequests := k8sModels.Requests{
		CPU:    conf.Env.Deployment.Resources.Requests.CPU,
		Memory: conf.Env.Deployment.Resources.Requests.Memory,
	}

	return &k8sModels.DeploymentPublic{
		ID:          "",
		Name:        name,
		Namespace:   namespace,
		DockerImage: dockerImage,
		EnvVars:     k8sEnvs,
		Resources: k8sModels.Resources{
			Limits:   defaultLimits,
			Requests: defaultRequests,
		},
	}
}

func createServicePublic(namespace, name string) *k8sModels.ServicePublic {
	port := conf.Env.Deployment.Port

	return &k8sModels.ServicePublic{
		ID:         "",
		Name:       name,
		Namespace:  namespace,
		Port:       port,
		TargetPort: port,
	}
}

func createIngressPublic(namespace, name string, serviceName string, servicePort int, domains []string) *k8sModels.IngressPublic {
	return &k8sModels.IngressPublic{
		ID:           "",
		Name:         name,
		Namespace:    namespace,
		ServiceName:  serviceName,
		ServicePort:  servicePort,
		IngressClass: conf.Env.Deployment.IngressClass,
		Hosts:        domains,
	}
}

func getExternalFQDN(name string) string {
	return fmt.Sprintf("%s.%s", name, conf.Env.Deployment.ParentDomain)
}

func CreateK8s(deploymentID string, userID string, params *deploymentModel.CreateParams) (*K8sResult, error) {
	log.Println("setting up k8s for", params.Name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup k8s for deployment %s. details: %s", params.Name, err)
	}

	deployment, err := deploymentModel.GetByID(deploymentID)
	if err != nil {
		return nil, makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", deploymentID, "not found for k8s setup assuming it was deleted")
		return nil, nil
	}

	zone := conf.Env.Deployment.GetZone(deployment.Zone)
	if zone == nil {
		return nil, fmt.Errorf("zone %s not found", deployment.Zone)
	}

	client, err := k8s.New(zone.Client)
	if err != nil {
		return nil, makeError(err)
	}

	ss := &deployment.Subsystems.K8s

	// Namespace
	namespace := &ss.Namespace
	if !ss.Namespace.Created() {
		namespace, err = createNamespace(client, deployment, createNamespacePublic(userID))
		if err != nil {
			return nil, makeError(err)
		}
	}

	// Deployment
	k8sDeployment := &ss.Deployment
	if !ss.Deployment.Created() {
		dockerRegistryProject := subsystemutils.GetPrefixedName(userID)
		dockerImage := fmt.Sprintf("%s/%s/%s", conf.Env.DockerRegistry.URL, dockerRegistryProject, deployment.Name)
		k8sDeployment, err = createK8sDeployment(client, deployment, createDeploymentPublic(namespace.FullName, deployment.Name, dockerImage, params.Envs))
		if err != nil {
			return nil, makeError(err)
		}
	}

	// Service
	service := &ss.Service
	if !ss.Service.Created() {
		service, err = createService(client, deployment, createServicePublic(namespace.FullName, deployment.Name))
		if err != nil {
			return nil, makeError(err)
		}
	}

	// Ingress
	ingress := &ss.Ingress
	if params.Private {
		if ss.Ingress.Created() {
			err = client.DeleteIngress(ss.Ingress.Namespace, ss.Ingress.Name)
			if err != nil {
				return nil, makeError(err)
			}
		}
		ingress = &k8sModels.IngressPublic{
			Placeholder: true,
		}

		err = deploymentModel.UpdateSubsystemByName(deployment.Name, "k8s", "ingress", ingress)
		if err != nil {
			return nil, makeError(err)
		}

	} else if !ss.Ingress.Created() {
		ingress, err = createIngress(client, deployment, createIngressPublic(
			namespace.FullName,
			deployment.Name,
			service.Name,
			service.Port,
			[]string{getExternalFQDN(deployment.Name)},
		))
		if err != nil {
			return nil, makeError(err)
		}
	}

	return &K8sResult{
		Namespace:  namespace,
		Deployment: k8sDeployment,
		Service:    service,
		Ingress:    ingress,
	}, nil
}

func DeleteK8s(name string) error {
	log.Println("deleting k8s for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s for deployment %s. details: %s", name, err)
	}

	deployment, err := deploymentModel.GetByName(name)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", name, "not found for k8s deletion. assuming it was deleted")
		return nil
	}

	zone := conf.Env.Deployment.GetZone(deployment.Zone)
	if zone == nil {
		return fmt.Errorf("zone %s not found", deployment.Zone)
	}

	client, err := k8s.New(zone.Client)
	if err != nil {
		return makeError(err)
	}

	ss := &deployment.Subsystems.K8s

	if ss.Ingress.Created() {
		err = client.DeleteIngress(ss.Ingress.Namespace, ss.Ingress.ID)
		if err != nil {
			return makeError(err)
		}

		err = deploymentModel.UpdateSubsystemByName(name, "k8s", "ingress", k8sModels.IngressPublic{})
		if err != nil {
			return makeError(err)
		}
	} else if ss.Ingress.Placeholder {
		err = deploymentModel.UpdateSubsystemByName(name, "k8s", "ingress", k8sModels.IngressPublic{})
		if err != nil {
			return makeError(err)
		}
	}

	if ss.Service.Created() {
		err = client.DeleteService(ss.Service.Namespace, ss.Service.ID)
		if err != nil {
			return makeError(err)
		}

		err = deploymentModel.UpdateSubsystemByName(name, "k8s", "service", k8sModels.ServicePublic{})
		if err != nil {
			return makeError(err)
		}
	}

	if ss.Deployment.Created() {
		err = client.DeleteDeployment(ss.Deployment.Namespace, ss.Deployment.ID)
		if err != nil {
			return makeError(err)
		}

		err = deploymentModel.UpdateSubsystemByName(name, "k8s", "deployment", k8sModels.DeploymentPublic{})
		if err != nil {
			return makeError(err)
		}
	}

	if ss.Namespace.Created() {
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
		log.Println("deployment", name, "not found for k8s update assuming it was deleted")
		return nil
	}

	zone := conf.Env.Deployment.GetZone(deployment.Zone)
	if zone == nil {
		return fmt.Errorf("zone %s not found", deployment.Zone)
	}

	client, err := k8s.New(zone.Client)
	if err != nil {
		return makeError(err)
	}

	if params.Envs != nil {
		if deployment.Subsystems.K8s.Deployment.Created() {
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

	if params.ExtraDomains != nil {
		if deployment.Subsystems.K8s.Ingress.Created() {
			ingress := deployment.Subsystems.K8s.Ingress
			if ingress.ID == "" {
				return nil
			}

			newPublic := &deployment.Subsystems.K8s.Ingress
			newPublic.Hosts = *params.ExtraDomains

			err = recreateIngress(client, deployment, newPublic)
			if err != nil {
				return makeError(err)
			}
		}
	}

	if params.Private != nil {
		emptyOrPlaceHolder := !deployment.Subsystems.K8s.Ingress.Created() || deployment.Subsystems.K8s.Ingress.Placeholder

		if *params.Private != emptyOrPlaceHolder {
			if !emptyOrPlaceHolder {
				err = client.DeleteIngress(deployment.Subsystems.K8s.Ingress.Namespace, deployment.Subsystems.K8s.Ingress.ID)
				if err != nil {
					return makeError(err)
				}

				err = deploymentModel.UpdateSubsystemByName(name, "k8s", "ingress", k8sModels.IngressPublic{})
				if err != nil {
					return makeError(err)
				}

				deployment, err = deploymentModel.GetByName(name)
				if err != nil {
					return makeError(err)
				}
			}

			if *params.Private {
				err = deploymentModel.UpdateSubsystemByName(name, "k8s", "ingress", k8sModels.IngressPublic{
					Placeholder: true,
				})
				if err != nil {
					return makeError(err)
				}
			} else {
				namespace := deployment.Subsystems.K8s.Namespace
				if !namespace.Created() {
					return nil
				}

				service := deployment.Subsystems.K8s.Service
				if !service.Created() {
					return nil
				}

				var domains []string
				if params.ExtraDomains == nil {
					domains = getAllDomainNames(deployment.Name, deployment.ExtraDomains)
				} else {
					domains = getAllDomainNames(deployment.Name, *params.ExtraDomains)
				}

				public := createIngressPublic(namespace.FullName, name, service.Name, service.Port, domains)
				_, err = createIngress(client, deployment, public)
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
	if !deployment.Subsystems.K8s.Deployment.Created() {
		return makeError(errors.New("can't restart deployment that is not yet created"))
	}

	zone := conf.Env.Deployment.GetZone(deployment.Zone)
	if zone == nil {
		return fmt.Errorf("zone %s not found", deployment.Zone)
	}

	client, err := k8s.New(zone.Client)
	if err != nil {
		return makeError(err)
	}

	err = client.RestartDeployment(&deployment.Subsystems.K8s.Deployment)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func RepairK8s(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair k8s %s. details: %s", name, err)
	}

	deployment, err := deploymentModel.GetByName(name)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", name, "not found when repairing k8s, assuming it was deleted")
		return nil
	}

	zone := conf.Env.Deployment.GetZone(deployment.Zone)
	if zone == nil {
		return fmt.Errorf("zone %s not found", deployment.Zone)
	}

	client, err := k8s.New(zone.Client)
	if err != nil {
		return makeError(err)
	}

	ss := deployment.Subsystems.K8s

	// temporary fix for missing resource limits and requests
	if ss.Deployment.Created() {
		res := &ss.Deployment.Resources
		if res.Limits.Memory == "" && res.Limits.CPU == "" && res.Requests.Memory == "" && res.Requests.CPU == "" {
			res.Limits.Memory = conf.Env.Deployment.Resources.Limits.Memory
			res.Limits.CPU = conf.Env.Deployment.Resources.Limits.CPU
			res.Requests.Memory = conf.Env.Deployment.Resources.Requests.Memory
			res.Requests.CPU = conf.Env.Deployment.Resources.Requests.CPU
		}
	}

	// namespace
	namespace, err := client.ReadNamespace(ss.Namespace.ID)
	if err != nil {
		return makeError(err)
	}

	if namespace == nil || !reflect.DeepEqual(ss.Namespace, *namespace) {
		log.Println("recreating namespace for deployment", name)
		err = recreateNamespace(client, deployment, &ss.Namespace)
		if err != nil {
			return makeError(err)
		}
	}

	// deployment
	k8sDeployment, err := client.ReadDeployment(ss.Namespace.FullName, ss.Deployment.ID)
	if err != nil {
		return makeError(err)
	}

	if k8sDeployment == nil || !reflect.DeepEqual(ss.Deployment, *k8sDeployment) {
		log.Println("recreating deployment for deployment", name)
		err = recreateK8sDeployment(client, deployment, &ss.Deployment)
		if err != nil {
			return makeError(err)
		}
	}

	// service
	service, err := client.ReadService(ss.Namespace.FullName, ss.Service.ID)
	if err != nil {
		return makeError(err)
	}

	if service == nil || !reflect.DeepEqual(ss.Service, *service) {
		log.Println("recreating service for deployment", name)
		err = recreateService(client, deployment, &ss.Service)
		if err != nil {
			return makeError(err)
		}
	}

	// ingress
	if deployment.Private != ss.Ingress.Placeholder {
		log.Println("recreating ingress for deployment due to mismatch with the private field", name)

		if deployment.Private {
			err = client.DeleteIngress(deployment.Subsystems.K8s.Ingress.Namespace, deployment.Subsystems.K8s.Ingress.ID)
			if err != nil {
				return makeError(err)
			}

			err = deploymentModel.UpdateSubsystemByName(name, "k8s", "ingress", k8sModels.IngressPublic{
				Placeholder: true,
			})
			if err != nil {
				return makeError(err)
			}
		} else {
			_, err = createIngress(client, deployment, createIngressPublic(
				deployment.Subsystems.K8s.Namespace.FullName,
				deployment.Name,
				deployment.Subsystems.K8s.Service.Name,
				deployment.Subsystems.K8s.Service.Port,
				getAllDomainNames(deployment.Name, deployment.ExtraDomains),
			))
			if err != nil {
				return makeError(err)
			}
		}
	} else if !ss.Ingress.Placeholder {
		ingress, err := client.ReadIngress(ss.Namespace.FullName, ss.Ingress.ID)
		if err != nil {
			return makeError(err)
		}

		if ingress == nil || !reflect.DeepEqual(ss.Ingress, *ingress) {
			log.Println("recreating ingress for deployment", name)
			err = recreateIngress(client, deployment, &ss.Ingress)
			if err != nil {
				return makeError(err)
			}
		}
	}

	return nil
}

func recreateNamespace(client *k8s.Client, deployment *deploymentModel.Deployment, newPublic *k8sModels.NamespacePublic) error {
	err := client.DeleteNamespace(deployment.Subsystems.K8s.Namespace.ID)
	if err != nil {
		return err
	}

	_, err = createNamespace(client, deployment, newPublic)
	if err != nil {
		return err
	}

	return nil
}

func recreateK8sDeployment(client *k8s.Client, deployment *deploymentModel.Deployment, newPublic *k8sModels.DeploymentPublic) error {
	err := client.DeleteDeployment(deployment.Subsystems.K8s.Namespace.FullName, deployment.Subsystems.K8s.Deployment.ID)
	if err != nil {
		return err
	}

	_, err = createK8sDeployment(client, deployment, newPublic)
	if err != nil {
		return err
	}

	return nil
}

func recreateService(client *k8s.Client, deployment *deploymentModel.Deployment, newPublic *k8sModels.ServicePublic) error {
	err := client.DeleteService(deployment.Subsystems.K8s.Namespace.FullName, deployment.Subsystems.K8s.Service.ID)
	if err != nil {
		return err
	}

	_, err = createService(client, deployment, newPublic)
	if err != nil {
		return err
	}

	return nil
}

func recreateIngress(client *k8s.Client, deployment *deploymentModel.Deployment, newPublic *k8sModels.IngressPublic) error {
	err := client.DeleteIngress(deployment.Subsystems.K8s.Namespace.FullName, deployment.Subsystems.K8s.Ingress.ID)
	if err != nil {
		return err
	}

	_, err = createIngress(client, deployment, newPublic)
	if err != nil {
		return err
	}

	return nil
}

func createNamespace(client *k8s.Client, deployment *deploymentModel.Deployment, public *k8sModels.NamespacePublic) (*k8sModels.NamespacePublic, error) {
	id, err := client.CreateNamespace(public)
	if err != nil {
		return nil, err
	}

	namespace, err := client.ReadNamespace(id)
	if err != nil {
		return nil, err
	}

	if namespace == nil {
		return nil, errors.New("failed to read namespace after creation")
	}

	err = deploymentModel.UpdateSubsystemByName(deployment.Name, "k8s", "namespace", namespace)
	if err != nil {
		return nil, err
	}

	deployment.Subsystems.K8s.Namespace = *namespace

	return namespace, nil
}

func createK8sDeployment(client *k8s.Client, deployment *deploymentModel.Deployment, public *k8sModels.DeploymentPublic) (*k8sModels.DeploymentPublic, error) {
	id, err := client.CreateDeployment(public)
	if err != nil {
		return nil, err
	}

	k8sDeployment, err := client.ReadDeployment(deployment.Subsystems.K8s.Namespace.FullName, id)
	if err != nil {
		return nil, err
	}

	if k8sDeployment == nil {
		return nil, errors.New("failed to read deployment after creation")
	}

	err = deploymentModel.UpdateSubsystemByName(deployment.Name, "k8s", "deployment", k8sDeployment)
	if err != nil {
		return nil, err
	}

	deployment.Subsystems.K8s.Deployment = *k8sDeployment

	return k8sDeployment, nil
}

func createService(client *k8s.Client, deployment *deploymentModel.Deployment, public *k8sModels.ServicePublic) (*k8sModels.ServicePublic, error) {
	id, err := client.CreateService(public)
	if err != nil {
		return nil, err
	}

	service, err := client.ReadService(public.Namespace, id)
	if err != nil {
		return nil, err
	}

	if service == nil {
		return nil, errors.New("failed to read service after creation")
	}

	err = deploymentModel.UpdateSubsystemByName(deployment.Name, "k8s", "service", service)
	if err != nil {
		return nil, err
	}

	deployment.Subsystems.K8s.Service = *service

	return service, nil
}

func createIngress(client *k8s.Client, deployment *deploymentModel.Deployment, public *k8sModels.IngressPublic) (*k8sModels.IngressPublic, error) {
	id, err := client.CreateIngress(public)
	if err != nil {
		return nil, err
	}

	ingress, err := client.ReadIngress(public.Namespace, id)
	if err != nil {
		return nil, err
	}

	if ingress == nil {
		return nil, errors.New("failed to read ingress after creation")
	}

	err = deploymentModel.UpdateSubsystemByName(deployment.Name, "k8s", "ingress", ingress)
	if err != nil {
		return nil, err
	}

	deployment.Subsystems.K8s.Ingress = *ingress

	return ingress, nil
}

func getAllDomainNames(name string, extraDomains []string) []string {
	domains := make([]string, len(extraDomains)+1)
	domains[0] = getExternalFQDN(name)
	copy(domains[1:], extraDomains)
	return domains
}

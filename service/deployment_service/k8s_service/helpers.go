package k8s_service

import (
	"errors"
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/models/sys/deployment/subsystems"
	"go-deploy/models/sys/enviroment"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/k8s"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"strconv"
)

type UpdateDbSubsystem func(string, string, string, interface{}) error

func createNamespacePublic(name string) *k8sModels.NamespacePublic {
	return &k8sModels.NamespacePublic{
		ID:       "",
		Name:     name,
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

func createServicePublic(namespace, name string, port *int) *k8sModels.ServicePublic {
	if port == nil {
		port = &conf.Env.Deployment.Port
	}

	return &k8sModels.ServicePublic{
		ID:         "",
		Name:       name,
		Namespace:  namespace,
		Port:       *port,
		TargetPort: *port,
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

func getExternalFQDN(name string, zone *enviroment.DeploymentZone) string {
	return fmt.Sprintf("%s.%s", name, zone.ParentDomain)
}

func getStorageManagerExternalFQDN(name string, zone *enviroment.DeploymentZone) string {
	return fmt.Sprintf("%s.%s", name, zone.StorageParentDomain)
}

func recreateNamespace(client *k8s.Client, id string, k8s *subsystems.K8s, newPublic *k8sModels.NamespacePublic, updateDb UpdateDbSubsystem) error {
	err := client.DeleteNamespace(k8s.Namespace.ID)
	if err != nil {
		return err
	}

	_, err = createNamespace(client, id, k8s, newPublic, updateDb)
	if err != nil {
		return err
	}

	return nil
}

func recreateK8sDeployment(client *k8s.Client, id string, k8s *subsystems.K8s, newPublic *k8sModels.DeploymentPublic, updateDb UpdateDbSubsystem) error {
	err := client.DeleteDeployment(k8s.Namespace.FullName, k8s.Deployment.ID)
	if err != nil {
		return err
	}

	_, err = createK8sDeployment(client, id, k8s, newPublic, updateDb)
	if err != nil {
		return err
	}

	return nil
}

func recreateService(client *k8s.Client, id string, k8s *subsystems.K8s, newPublic *k8sModels.ServicePublic, updateDb UpdateDbSubsystem) error {
	err := client.DeleteService(k8s.Namespace.FullName, k8s.Service.ID)
	if err != nil {
		return err
	}

	_, err = createService(client, id, k8s, newPublic, updateDb)
	if err != nil {
		return err
	}

	return nil
}

func recreateIngress(client *k8s.Client, id string, k8s *subsystems.K8s, newPublic *k8sModels.IngressPublic, updateDb UpdateDbSubsystem) error {
	err := client.DeleteIngress(k8s.Namespace.FullName, k8s.Ingress.ID)
	if err != nil {
		return err
	}

	_, err = createIngress(client, id, k8s, newPublic, updateDb)
	if err != nil {
		return err
	}

	return nil
}

func createNamespace(client *k8s.Client, id string, k8s *subsystems.K8s, public *k8sModels.NamespacePublic, updateDb UpdateDbSubsystem) (*k8sModels.NamespacePublic, error) {
	createdID, err := client.CreateNamespace(public)
	if err != nil {
		return nil, err
	}

	namespace, err := client.ReadNamespace(createdID)
	if err != nil {
		return nil, err
	}

	if namespace == nil {
		return nil, errors.New("failed to read namespace after creation")
	}

	err = updateDb(id, "k8s", "namespace", namespace)
	if err != nil {
		return nil, err
	}

	k8s.Namespace = *namespace

	return namespace, nil
}

func createK8sDeployment(client *k8s.Client, id string, k8s *subsystems.K8s, public *k8sModels.DeploymentPublic, updateDb UpdateDbSubsystem) (*k8sModels.DeploymentPublic, error) {
	createdID, err := client.CreateDeployment(public)
	if err != nil {
		return nil, err
	}

	k8sDeployment, err := client.ReadDeployment(k8s.Namespace.FullName, createdID)
	if err != nil {
		return nil, err
	}

	if k8sDeployment == nil {
		return nil, errors.New("failed to read deployment after creation")
	}

	err = updateDb(id, "k8s", "deployment", k8sDeployment)
	if err != nil {
		return nil, err
	}

	k8s.Deployment = *k8sDeployment

	return k8sDeployment, nil
}

func createService(client *k8s.Client, id string, k8s *subsystems.K8s, public *k8sModels.ServicePublic, updateDb UpdateDbSubsystem) (*k8sModels.ServicePublic, error) {
	createdID, err := client.CreateService(public)
	if err != nil {
		return nil, err
	}

	service, err := client.ReadService(public.Namespace, createdID)
	if err != nil {
		return nil, err
	}

	if service == nil {
		return nil, errors.New("failed to read service after creation")
	}

	err = updateDb(id, "k8s", "service", service)
	if err != nil {
		return nil, err
	}

	k8s.Service = *service

	return service, nil
}

func createIngress(client *k8s.Client, id string, k8s *subsystems.K8s, public *k8sModels.IngressPublic, updateDb UpdateDbSubsystem) (*k8sModels.IngressPublic, error) {
	createdID, err := client.CreateIngress(public)
	if err != nil {
		return nil, err
	}

	ingress, err := client.ReadIngress(public.Namespace, createdID)
	if err != nil {
		return nil, err
	}

	if ingress == nil {
		return nil, errors.New("failed to read ingress after creation")
	}

	err = updateDb(id, "k8s", "ingress", ingress)
	if err != nil {
		return nil, err
	}

	k8s.Ingress = *ingress

	return ingress, nil
}

func getAllDomainNames(name string, extraDomains []string, zone *enviroment.DeploymentZone) []string {
	domains := make([]string, len(extraDomains)+1)
	domains[0] = getExternalFQDN(name, zone)
	copy(domains[1:], extraDomains)
	return domains
}

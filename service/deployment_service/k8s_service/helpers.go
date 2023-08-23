package k8s_service

import (
	"errors"
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/models/sys/deployment/storage_manager"
	"go-deploy/models/sys/deployment/subsystems"
	"go-deploy/models/sys/enviroment"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/k8s"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"log"
	"reflect"
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

func createDeploymentPublic(namespace, name, dockerImage string, envs []deploymentModel.Env, volumes []deploymentModel.Volume, initCommands []string) *k8sModels.DeploymentPublic {
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

	k8sVolumes := make([]k8sModels.Volume, len(volumes))
	for i, volume := range volumes {
		k8sVolumes[i] = k8sModels.Volume{
			Name:      volume.Name,
			PvcName:   fmt.Sprintf("%s-%s", name, volume.Name),
			MountPath: volume.AppPath,
			Init:      volume.Init,
		}
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
		Command:        nil,
		Args:           nil,
		InitCommands:   initCommands,
		InitContainers: nil,
		Volumes:        k8sVolumes,
	}
}

func createStorageManagerDeploymentPublic(namespace, name string, volumes []storage_manager.Volume, initCommands []string) *k8sModels.DeploymentPublic {
	k8sVolumes := make([]k8sModels.Volume, len(volumes))
	for i, volume := range volumes {
		k8sVolumes[i] = k8sModels.Volume{
			Name:      volume.Name,
			PvcName:   volume.Name,
			MountPath: volume.AppPath,
			Init:      volume.Init,
		}
	}

	defaultLimits := k8sModels.Limits{
		CPU:    conf.Env.Deployment.Resources.Limits.CPU,
		Memory: conf.Env.Deployment.Resources.Limits.Memory,
	}

	defaultRequests := k8sModels.Requests{
		CPU:    conf.Env.Deployment.Resources.Requests.CPU,
		Memory: conf.Env.Deployment.Resources.Requests.Memory,
	}

	args := []string{
		"--noauth",
		"--root=/deploy",
		"--database=/data/database.db",
		"--port=80",
	}

	return &k8sModels.DeploymentPublic{
		ID:          "",
		Name:        name,
		Namespace:   namespace,
		DockerImage: "filebrowser/filebrowser",
		EnvVars:     nil,
		Resources: k8sModels.Resources{
			Limits:   defaultLimits,
			Requests: defaultRequests,
		},
		Command:        nil,
		Args:           args,
		InitCommands:   initCommands,
		InitContainers: nil,
		Volumes:        k8sVolumes,
	}
}

func createServicePublic(namespace, name string, port int) *k8sModels.ServicePublic {
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

func createPvPublic(name string, capacity, nfsPath, nfsServer string) *k8sModels.PvPublic {
	return &k8sModels.PvPublic{
		ID:        "",
		Name:      name,
		Capacity:  capacity,
		NfsPath:   nfsPath,
		NfsServer: nfsServer,
	}
}

func createPvcPublic(namespace, name, capacity, pvName string) *k8sModels.PvcPublic {
	return &k8sModels.PvcPublic{
		ID:        "",
		Name:      name,
		Namespace: namespace,
		Capacity:  capacity,
		PvName:    pvName,
	}
}

func createJobPublic(namespace, name, image string, command, args []string, volumes []storage_manager.Volume) *k8sModels.JobPublic {
	k8sVolumes := make([]k8sModels.Volume, len(volumes))
	for i, volume := range volumes {
		k8sVolumes[i] = k8sModels.Volume{
			Name:      volume.Name,
			PvcName:   volume.Name,
			MountPath: volume.AppPath,
			Init:      volume.Init,
		}
	}

	return &k8sModels.JobPublic{
		ID:        "",
		Name:      name,
		Namespace: namespace,
		Image:     image,
		Command:   command,
		Args:      args,
		Volumes:   k8sVolumes,
	}
}

func getExternalFQDN(name string, zone *enviroment.DeploymentZone) string {
	return fmt.Sprintf("%s.%s", name, zone.ParentDomain)
}

func getStorageManagerExternalFQDN(name string, zone *enviroment.DeploymentZone) string {
	return fmt.Sprintf("%s.%s", name, zone.Storage.ParentDomain)
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

func recreatePV(client *k8s.Client, id, name string, k8s *subsystems.K8s, newPublic *k8sModels.PvPublic, updateDb UpdateDbSubsystem) error {
	pv, ok := k8s.PvMap[name]
	if ok {
		err := client.DeletePV(pv.ID)
		if err != nil {
			return err
		}
	}

	_, err := createPV(client, id, name, k8s, newPublic, updateDb)
	if err != nil {
		return err
	}

	return nil
}

func recreatePVC(client *k8s.Client, id, name string, k8s *subsystems.K8s, newPublic *k8sModels.PvcPublic, updateDb UpdateDbSubsystem) error {
	pvc, ok := k8s.PvcMap[name]
	if ok {
		err := client.DeletePVC(pvc.Namespace, pvc.ID)
		if err != nil {
			return err
		}
	}

	_, err := createPVC(client, id, name, k8s, newPublic, updateDb)
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

func createPV(client *k8s.Client, id, name string, k8s *subsystems.K8s, public *k8sModels.PvPublic, updateDb UpdateDbSubsystem) (*k8sModels.PvPublic, error) {
	createdID, err := client.CreatePV(public)
	if err != nil {
		return nil, err
	}

	pv, err := client.ReadPV(createdID)
	if err != nil {
		return nil, err
	}

	if pv == nil {
		return nil, errors.New("failed to read persistent volume after creation")
	}

	newPvMap := make(map[string]k8sModels.PvPublic)
	for k, v := range k8s.PvMap {
		newPvMap[k] = v
	}
	newPvMap[name] = *pv

	err = updateDb(id, "k8s", "pvMap", newPvMap)
	if err != nil {
		return nil, err
	}

	k8s.PvMap[name] = *pv

	return pv, nil
}

func createPVC(client *k8s.Client, id, name string, k8s *subsystems.K8s, public *k8sModels.PvcPublic, updateDb UpdateDbSubsystem) (*k8sModels.PvcPublic, error) {
	createdID, err := client.CreatePVC(public)
	if err != nil {
		return nil, err
	}

	pvc, err := client.ReadPVC(public.Namespace, createdID)
	if err != nil {
		return nil, err
	}

	if pvc == nil {
		return nil, errors.New("failed to read persistent volume claim after creation")
	}

	newPvcMap := make(map[string]k8sModels.PvcPublic)
	for k, v := range k8s.PvcMap {
		newPvcMap[k] = v
	}
	newPvcMap[name] = *pvc

	err = updateDb(id, "k8s", "pvcMap", newPvcMap)
	if err != nil {
		return nil, err
	}

	k8s.PvcMap[name] = *pvc

	return pvc, nil
}

func createJob(client *k8s.Client, id, name string, k8s *subsystems.K8s, public *k8sModels.JobPublic, updateDb UpdateDbSubsystem) (*k8sModels.JobPublic, error) {
	createdID, err := client.CreateJob(public)
	if err != nil {
		return nil, err
	}

	job, err := client.ReadJob(public.Namespace, createdID)
	if err != nil {
		return nil, err
	}

	if job == nil {
		return nil, errors.New("failed to read job after creation")
	}

	newJobMap := make(map[string]k8sModels.JobPublic)
	for k, v := range k8s.JobMap {
		newJobMap[k] = v
	}
	newJobMap[name] = *job

	err = updateDb(id, "k8s", "jobMap", newJobMap)
	if err != nil {
		return nil, err
	}

	k8s.JobMap[name] = *job

	return job, nil
}

func deleteDeployment(client *k8s.Client, id string, k8s *subsystems.K8s, updateDb UpdateDbSubsystem) error {
	err := client.DeleteDeployment(k8s.Namespace.FullName, k8s.Deployment.ID)
	if err != nil {
		return err
	}

	err = updateDb(id, "k8s", "deployment", nil)
	if err != nil {
		return err
	}

	k8s.Deployment = k8sModels.DeploymentPublic{}

	return nil
}

func deleteService(client *k8s.Client, id string, k8s *subsystems.K8s, updateDb UpdateDbSubsystem) error {
	err := client.DeleteService(k8s.Namespace.FullName, k8s.Service.ID)
	if err != nil {
		return err
	}

	err = updateDb(id, "k8s", "service", nil)
	if err != nil {
		return err
	}

	k8s.Service = k8sModels.ServicePublic{}

	return nil
}

func deleteIngress(client *k8s.Client, id string, k8s *subsystems.K8s, updateDb UpdateDbSubsystem) error {
	err := client.DeleteIngress(k8s.Namespace.FullName, k8s.Ingress.ID)
	if err != nil {
		return err
	}

	err = updateDb(id, "k8s", "ingress", nil)
	if err != nil {
		return err
	}

	k8s.Ingress = k8sModels.IngressPublic{}

	return nil
}

func deletePV(client *k8s.Client, id, name string, k8s *subsystems.K8s, updateDb UpdateDbSubsystem) error {
	pv, ok := k8s.PvMap[name]
	if ok {
		err := client.DeletePV(pv.ID)
		if err != nil {
			return err
		}
	}

	newPvMap := make(map[string]k8sModels.PvPublic)
	for k, v := range k8s.PvMap {
		if k != name {
			newPvMap[k] = v
		}
	}

	err := updateDb(id, "k8s", "pvMap", newPvMap)
	if err != nil {
		return err
	}

	k8s.PvMap = newPvMap

	return nil
}

func deletePVC(client *k8s.Client, id, name string, k8s *subsystems.K8s, updateDb UpdateDbSubsystem) error {
	pvc, ok := k8s.PvcMap[name]
	if ok {
		err := client.DeletePVC(pvc.Namespace, pvc.ID)
		if err != nil {
			return err
		}
	}

	newPvcMap := make(map[string]k8sModels.PvcPublic)
	for k, v := range k8s.PvcMap {
		if k != name {
			newPvcMap[k] = v
		}
	}

	err := updateDb(id, "k8s", "pvcMap", newPvcMap)
	if err != nil {
		return err
	}

	k8s.PvcMap = newPvcMap

	return nil
}

func deleteJob(client *k8s.Client, id, name string, k8s *subsystems.K8s, updateDb UpdateDbSubsystem) error {
	job, ok := k8s.JobMap[name]
	if ok {
		err := client.DeleteJob(job.Namespace, job.ID)
		if err != nil {
			return err
		}
	}

	newJobMap := make(map[string]k8sModels.JobPublic)
	for k, v := range k8s.JobMap {
		if k != name {
			newJobMap[k] = v
		}
	}

	err := updateDb(id, "k8s", "jobMap", newJobMap)
	if err != nil {
		return err
	}

	k8s.JobMap = newJobMap

	return nil
}

func repairDeployment(client *k8s.Client, id string, k8s *subsystems.K8s, updateDb UpdateDbSubsystem) error {
	deployment, err := client.ReadDeployment(k8s.Namespace.FullName, k8s.Deployment.ID)
	if err != nil {
		return err
	}

	if deployment == nil || !reflect.DeepEqual(k8s.Deployment, *deployment) {
		log.Println("recreating deployment for deployment", id)
		err = recreateK8sDeployment(client, id, k8s, &k8s.Deployment, updateDb)
		if err != nil {
			return err
		}
	}

	return nil
}

func repairService(client *k8s.Client, id string, k8s *subsystems.K8s, updateDb UpdateDbSubsystem) error {
	service, err := client.ReadService(k8s.Namespace.FullName, k8s.Service.ID)
	if err != nil {
		return err
	}

	if service == nil || !reflect.DeepEqual(k8s.Service, *service) {
		log.Println("recreating service for storage manager", id)
		err = recreateService(client, id, k8s, &k8s.Service, updateDb)
		if err != nil {
			return err
		}
	}

	return nil
}

func repairIngress(client *k8s.Client, id string, k8s *subsystems.K8s, updateDb UpdateDbSubsystem) error {
	ingress, err := client.ReadIngress(k8s.Namespace.FullName, k8s.Ingress.ID)
	if err != nil {
		return err
	}

	if ingress == nil || !reflect.DeepEqual(k8s.Ingress, *ingress) {
		log.Println("recreating ingress for storage manager", id)
		err = recreateIngress(client, id, k8s, &k8s.Ingress, updateDb)
		if err != nil {
			return err
		}
	}

	return nil
}

func getAllDomainNames(name string, extraDomains []string, zone *enviroment.DeploymentZone) []string {
	domains := make([]string, len(extraDomains)+1)
	domains[0] = getExternalFQDN(name, zone)
	copy(domains[1:], extraDomains)
	return domains
}

package k8s_service

import (
	"errors"
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/models/sys/deployment/storage_manager"
	"go-deploy/models/sys/deployment/subsystems"
	"go-deploy/models/sys/enviroment"
	userModel "go-deploy/models/sys/user"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/k8s"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/utils"
	"go-deploy/utils/subsystemutils"
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

func createMainAppDeploymentPublic(namespace, name, userID string, envs []deploymentModel.Env, volumes []deploymentModel.Volume, initCommands []string) *k8sModels.DeploymentPublic {
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
		pvcName := fmt.Sprintf("%s-%s", name, volume.Name)
		k8sVolumes[i] = k8sModels.Volume{
			Name:      volume.Name,
			PvcName:   &pvcName,
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

	dockerRegistryProject := subsystemutils.GetPrefixedName(userID)
	dockerImage := fmt.Sprintf("%s/%s/%s", conf.Env.DockerRegistry.URL, dockerRegistryProject, name)

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

func createFileBrowserDeploymentPublic(namespace, name string, volumes []storage_manager.Volume, initCommands []string) *k8sModels.DeploymentPublic {
	k8sVolumes := make([]k8sModels.Volume, len(volumes))
	for i, volume := range volumes {
		pvcName := volume.Name
		k8sVolumes[i] = k8sModels.Volume{
			Name:      volume.Name,
			PvcName:   &pvcName,
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

func createOAuthProxyDeploymentPublic(namespace, name, userID string, zone *enviroment.DeploymentZone) *k8sModels.DeploymentPublic {

	defaultLimits := k8sModels.Limits{
		CPU:    conf.Env.Deployment.Resources.Limits.CPU,
		Memory: conf.Env.Deployment.Resources.Limits.Memory,
	}

	defaultRequests := k8sModels.Requests{
		CPU:    conf.Env.Deployment.Resources.Requests.CPU,
		Memory: conf.Env.Deployment.Resources.Requests.Memory,
	}

	user, err := userModel.New().GetByID(userID)
	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("failed to get user by id when creating oauth proxy deployment public. details: %w", err))
		return nil
	}

	volumes := []k8sModels.Volume{
		{
			Name:      "oauth-proxy-config",
			PvcName:   nil,
			MountPath: "/mnt/config",
			Init:      true,
		},
		{
			Name:      "oauth-proxy-config",
			PvcName:   nil,
			MountPath: "/mnt",
			Init:      false,
		},
	}

	issuer := conf.Env.Keycloak.Url + "/realms/" + conf.Env.Keycloak.Realm
	redirectURL := fmt.Sprintf("https://%s.%s/oauth2/callback", userID, zone.Storage.ParentDomain)
	upstream := "http://storage-manager"

	args := []string{
		"--http-address=0.0.0.0:4180",
		"--reverse-proxy=true",
		"--provider=oidc",
		"--redirect-url=" + redirectURL,
		"--oidc-issuer-url=" + issuer,
		"--cookie-expire=168h",
		"--cookie-refresh=1h",
		"--pass-authorization-header=true",
		"--scope=openid email",
		"--upstream=" + upstream,
		"--client-id=" + conf.Env.Keycloak.StorageClient.ClientID,
		"--client-secret=" + conf.Env.Keycloak.StorageClient.ClientSecret,
		"--cookie-secret=qHKgjlAFQBZOnGcdH5jIKV0Auzx5r8jzZenxhJnlZJg=",
		"--cookie-secure=true",
		"--ssl-insecure-skip-verify=true",
		"--insecure-oidc-allow-unverified-email=true",
		"--skip-provider-button=true",
		"--pass-authorization-header=true",
		"--ssl-upstream-insecure-skip-verify=true",
		//"--session-store-type=redis",
		//"--redis-connection-url=redis://redis-master:6379",
		"--code-challenge-method=S256",
		"--oidc-groups-claim=groups",
		"--allowed-group=" + conf.Env.Keycloak.AdminGroup,
		"--authenticated-emails-file=/mnt/authenticated-emails-list",
	}

	initContainers := []k8sModels.InitContainer{
		{
			Name:    "oauth-proxy-config-init",
			Image:   "busybox",
			Command: []string{"sh", "-c", fmt.Sprintf("mkdir -p /mnt/config && echo %s > /mnt/config/authenticated-emails-list", user.Email)},
			Args:    nil,
		},
	}

	return &k8sModels.DeploymentPublic{
		ID:          "",
		Name:        name,
		Namespace:   namespace,
		DockerImage: "quay.io/oauth2-proxy/oauth2-proxy:latest",
		EnvVars:     nil,
		Resources: k8sModels.Resources{
			Limits:   defaultLimits,
			Requests: defaultRequests,
		},
		Command:        nil,
		Args:           args,
		InitCommands:   nil,
		InitContainers: initContainers,
		Volumes:        volumes,
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
			PvcName:   &volume.Name,
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
	err := deleteNamespace(client, id, k8s, updateDb)
	if err != nil {
		return err
	}

	_, err = createNamespace(client, id, k8s, newPublic, updateDb)
	if err != nil {
		return err
	}

	return nil
}

func recreateK8sDeployment(client *k8s.Client, id, name string, k8s *subsystems.K8s, newPublic *k8sModels.DeploymentPublic, updateDb UpdateDbSubsystem) error {
	err := deleteK8sDeployment(client, id, name, k8s, updateDb)
	if err != nil {
		return err
	}

	_, err = createK8sDeployment(client, id, name, k8s, newPublic, updateDb)
	if err != nil {
		return err
	}

	return nil
}

func recreateService(client *k8s.Client, id, name string, k8s *subsystems.K8s, newPublic *k8sModels.ServicePublic, updateDb UpdateDbSubsystem) error {
	err := deleteService(client, id, name, k8s, updateDb)
	if err != nil {
		return err
	}

	_, err = createService(client, id, name, k8s, newPublic, updateDb)
	if err != nil {
		return err
	}

	return nil
}

func recreateIngress(client *k8s.Client, id, name string, k8s *subsystems.K8s, newPublic *k8sModels.IngressPublic, updateDb UpdateDbSubsystem) error {
	err := deleteIngress(client, id, name, k8s, updateDb)
	if err != nil {
		return err
	}

	_, err = createIngress(client, id, name, k8s, newPublic, updateDb)
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

func createK8sDeployment(client *k8s.Client, id, name string, k8s *subsystems.K8s, public *k8sModels.DeploymentPublic, updateDb UpdateDbSubsystem) (*k8sModels.DeploymentPublic, error) {
	createdID, err := client.CreateDeployment(public)
	if err != nil {
		return nil, err
	}

	deployment, err := client.ReadDeployment(public.Namespace, createdID)
	if err != nil {
		return nil, err
	}

	if deployment == nil {
		log.Printf("failed to read deployment after creation. assuming it was deleted")
		return nil, nil
	}

	newMap := make(map[string]k8sModels.DeploymentPublic)
	for k, v := range k8s.DeploymentMap {
		newMap[k] = v
	}
	newMap[name] = *deployment

	err = updateDb(id, "k8s", "deploymentMap", newMap)
	if err != nil {
		return nil, err
	}

	k8s.DeploymentMap = newMap

	return deployment, nil
}

func createService(client *k8s.Client, id, name string, k8s *subsystems.K8s, public *k8sModels.ServicePublic, updateDb UpdateDbSubsystem) (*k8sModels.ServicePublic, error) {
	createdID, err := client.CreateService(public)
	if err != nil {
		return nil, err
	}

	service, err := client.ReadService(public.Namespace, createdID)
	if err != nil {
		return nil, err
	}

	if service == nil {
		log.Printf("failed to read service after creation. assuming it was deleted")
		return nil, nil
	}

	newMap := make(map[string]k8sModels.ServicePublic)
	for k, v := range k8s.ServiceMap {
		newMap[k] = v
	}
	newMap[name] = *service

	err = updateDb(id, "k8s", "serviceMap", newMap)
	if err != nil {
		return nil, err
	}

	k8s.ServiceMap = newMap

	return service, nil
}

func createIngress(client *k8s.Client, id, name string, k8s *subsystems.K8s, public *k8sModels.IngressPublic, updateDb UpdateDbSubsystem) (*k8sModels.IngressPublic, error) {
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

	newMap := make(map[string]k8sModels.IngressPublic)
	for k, v := range k8s.IngressMap {
		newMap[k] = v
	}
	newMap[name] = *ingress

	err = updateDb(id, "k8s", "ingressMap", newMap)
	if err != nil {
		return nil, err
	}

	k8s.IngressMap = newMap

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

	newMap := make(map[string]k8sModels.PvPublic)
	for k, v := range k8s.PvMap {
		newMap[k] = v
	}
	newMap[name] = *pv

	err = updateDb(id, "k8s", "pvMap", newMap)
	if err != nil {
		return nil, err
	}

	k8s.PvMap = newMap

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

	newMap := make(map[string]k8sModels.PvcPublic)
	for k, v := range k8s.PvcMap {
		newMap[k] = v
	}
	newMap[name] = *pvc

	err = updateDb(id, "k8s", "pvcMap", newMap)
	if err != nil {
		return nil, err
	}

	k8s.PvcMap = newMap

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

	newMap := make(map[string]k8sModels.JobPublic)
	for k, v := range k8s.JobMap {
		newMap[k] = v
	}
	newMap[name] = *job

	err = updateDb(id, "k8s", "jobMap", newMap)
	if err != nil {
		return nil, err
	}

	k8s.JobMap = newMap

	return job, nil
}

func deleteNamespace(client *k8s.Client, id string, k8s *subsystems.K8s, updateDb UpdateDbSubsystem) error {
	// never actually deleted the namespace to prevent race conditions

	err := updateDb(id, "k8s", "namespace", nil)
	if err != nil {
		return err
	}

	k8s.Namespace = k8sModels.NamespacePublic{}

	return nil
}

func deleteK8sDeployment(client *k8s.Client, id, name string, k8s *subsystems.K8s, updateDb UpdateDbSubsystem) error {
	deployment, ok := k8s.DeploymentMap[name]
	if ok {
		err := client.DeleteDeployment(k8s.Namespace.FullName, deployment.ID)
		if err != nil {
			return err
		}
	}

	newMap := make(map[string]k8sModels.DeploymentPublic)
	for k, v := range k8s.DeploymentMap {
		if k != name {
			newMap[k] = v
		}
	}

	err := updateDb(id, "k8s", "deploymentMap", newMap)
	if err != nil {
		return err
	}

	k8s.DeploymentMap = newMap

	return nil
}

func deleteService(client *k8s.Client, id, name string, k8s *subsystems.K8s, updateDb UpdateDbSubsystem) error {
	service, ok := k8s.ServiceMap[name]
	if ok {
		err := client.DeleteService(k8s.Namespace.FullName, service.ID)
		if err != nil {
			return err
		}
	}

	newMap := make(map[string]k8sModels.ServicePublic)
	for k, v := range k8s.ServiceMap {
		if k != name {
			newMap[k] = v
		}
	}

	err := updateDb(id, "k8s", "serviceMap", nil)
	if err != nil {
		return err
	}

	k8s.ServiceMap = newMap

	return nil
}

func deleteIngress(client *k8s.Client, id, name string, k8s *subsystems.K8s, updateDb UpdateDbSubsystem) error {
	ingress, ok := k8s.IngressMap[name]
	if ok {
		err := client.DeleteIngress(k8s.Namespace.FullName, ingress.ID)
		if err != nil {
			return err
		}
	}

	newMap := make(map[string]k8sModels.IngressPublic)
	for k, v := range k8s.IngressMap {
		if k != name {
			newMap[k] = v
		}
	}

	err := updateDb(id, "k8s", "ingressMap", nil)
	if err != nil {
		return err
	}

	k8s.IngressMap = newMap

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

	newMap := make(map[string]k8sModels.PvPublic)
	for k, v := range k8s.PvMap {
		if k != name {
			newMap[k] = v
		}
	}

	err := updateDb(id, "k8s", "pvMap", newMap)
	if err != nil {
		return err
	}

	k8s.PvMap = newMap

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

	newMap := make(map[string]k8sModels.PvcPublic)
	for k, v := range k8s.PvcMap {
		if k != name {
			newMap[k] = v
		}
	}

	err := updateDb(id, "k8s", "pvcMap", newMap)
	if err != nil {
		return err
	}

	k8s.PvcMap = newMap

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

	newMap := make(map[string]k8sModels.JobPublic)
	for k, v := range k8s.JobMap {
		if k != name {
			newMap[k] = v
		}
	}

	err := updateDb(id, "k8s", "jobMap", newMap)
	if err != nil {
		return err
	}

	k8s.JobMap = newMap

	return nil
}

func repairDeployment(client *k8s.Client, id, name string, k8s *subsystems.K8s, updateDb UpdateDbSubsystem) error {
	dbDeployment, ok := k8s.DeploymentMap[name]
	if !ok {
		return nil
	}

	installedDeployment, err := client.ReadDeployment(k8s.Namespace.FullName, dbDeployment.ID)
	if err != nil {
		return err
	}

	if installedDeployment == nil {
		return nil
	}

	if !ok || !reflect.DeepEqual(dbDeployment, *installedDeployment) {
		log.Println("recreating deployment for deployment", id)
		err = recreateK8sDeployment(client, id, name, k8s, &dbDeployment, updateDb)
		if err != nil {
			return err
		}
	}

	return nil
}

func repairService(client *k8s.Client, id, name string, k8s *subsystems.K8s, updateDb UpdateDbSubsystem) error {
	dbService, ok := k8s.ServiceMap[name]
	if !ok {
		return nil
	}

	installedService, err := client.ReadService(k8s.Namespace.FullName, dbService.ID)
	if err != nil {
		return err
	}

	if installedService == nil {
		return nil
	}

	if !ok || !reflect.DeepEqual(dbService, *installedService) {
		log.Println("recreating service for deployment", id)
		err = recreateService(client, id, name, k8s, &dbService, updateDb)
		if err != nil {
			return err
		}
	}

	return nil
}

func repairIngress(client *k8s.Client, id, name string, k8s *subsystems.K8s, updateDb UpdateDbSubsystem) error {
	dbIngress, ok := k8s.IngressMap[name]
	if !ok {
		return nil
	}

	installedService, err := client.ReadIngress(k8s.Namespace.FullName, dbIngress.ID)
	if err != nil {
		return err
	}

	if installedService == nil {
		return nil
	}

	if !ok || !reflect.DeepEqual(dbIngress, *installedService) {
		log.Println("recreating service for deployment", id)
		err = recreateIngress(client, id, name, k8s, &dbIngress, updateDb)
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

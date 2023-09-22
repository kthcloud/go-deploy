package helpers

import (
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/models/sys/deployment/storage_manager"
	"go-deploy/models/sys/enviroment"
	userModel "go-deploy/models/sys/user"
	"go-deploy/pkg/conf"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/utils"
	"strconv"
)

func CreateNamespacePublic(name string) *k8sModels.NamespacePublic {
	return &k8sModels.NamespacePublic{
		ID:       "",
		Name:     name,
		FullName: "",
	}
}

func CreateMainAppDeploymentPublic(namespace, name, image string, port int, envs []deploymentModel.Env, volumes []deploymentModel.Volume, initCommands []string) *k8sModels.DeploymentPublic {
	k8sEnvs := []k8sModels.EnvVar{
		{Name: "PORT", Value: strconv.Itoa(port)},
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

	return &k8sModels.DeploymentPublic{
		ID:          "",
		Name:        name,
		Namespace:   namespace,
		DockerImage: image,
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

func CreateFileBrowserDeploymentPublic(namespace, name string, volumes []storage_manager.Volume, initCommands []string) *k8sModels.DeploymentPublic {
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

func CreateOAuthProxyDeploymentPublic(namespace, name, userID string, zone *enviroment.DeploymentZone) *k8sModels.DeploymentPublic {

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

func CreateServicePublic(namespace, name string, externalPort, internalPort int) *k8sModels.ServicePublic {
	return &k8sModels.ServicePublic{
		ID:         "",
		Name:       name,
		Namespace:  namespace,
		Port:       externalPort,
		TargetPort: internalPort,
	}
}

func CreateIngressPublic(namespace, name string, serviceName string, servicePort int, domains []string) *k8sModels.IngressPublic {
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

func CreatePvPublic(name string, capacity, nfsPath, nfsServer string) *k8sModels.PvPublic {
	return &k8sModels.PvPublic{
		ID:        "",
		Name:      name,
		Capacity:  capacity,
		NfsPath:   nfsPath,
		NfsServer: nfsServer,
	}
}

func CreatePvcPublic(namespace, name, capacity, pvName string) *k8sModels.PvcPublic {
	return &k8sModels.PvcPublic{
		ID:        "",
		Name:      name,
		Namespace: namespace,
		Capacity:  capacity,
		PvName:    pvName,
	}
}

func CreateJobPublic(namespace, name, image string, command, args []string, volumes []storage_manager.Volume) *k8sModels.JobPublic {
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

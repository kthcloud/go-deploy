package resources

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go-deploy/models/sys/deployment"
	storageManagerModel "go-deploy/models/sys/deployment/storage_manager"
	"go-deploy/models/sys/enviroment"
	userModel "go-deploy/models/sys/user"
	"go-deploy/models/sys/vm"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service"
	"go-deploy/service/constants"
	"go-deploy/utils"
	"golang.org/x/exp/slices"
	v1 "k8s.io/api/core/v1"
	"path"
)

type K8sGenerator struct {
	*PublicGeneratorType
	namespace string
	client    *k8s.Client
}

func (kg *K8sGenerator) Namespace() *models.NamespacePublic {
	if kg.d.deployment != nil {
		if ns := &kg.d.deployment.Subsystems.K8s.Namespace; service.Created(ns) {
			return ns
		} else {
			return &models.NamespacePublic{
				Name: kg.namespace,
			}
		}
	}

	if kg.v.vm != nil {
		createNamespace := false
		for _, port := range kg.v.vm.Ports {
			if port.HttpProxy != nil {
				createNamespace = true
				break
			}
		}

		if !createNamespace {
			return nil
		}

		if ns := &kg.v.vm.Subsystems.K8s.Namespace; service.Created(ns) {
			return ns
		} else {
			return &models.NamespacePublic{
				Name: kg.namespace,
			}
		}
	}

	if kg.s.storageManager != nil {
		if ns := &kg.s.storageManager.Subsystems.K8s.Namespace; service.Created(ns) {
			return ns
		} else {
			return &models.NamespacePublic{
				Name: kg.namespace,
			}
		}
	}

	return nil
}

func (kg *K8sGenerator) Deployments() []models.DeploymentPublic {
	var res []models.DeploymentPublic

	if kg.d.deployment != nil {
		if k8sDeployment := kg.d.deployment.Subsystems.K8s.GetDeployment(kg.d.deployment.Name); service.Created(k8sDeployment) {
			mainApp := kg.d.deployment.GetMainApp()

			var volumes []models.Volume
			for _, volume := range mainApp.Volumes {
				pvcName := getDeploymentPvcName(kg.d.deployment, volume.Name)
				volumes = append(volumes, models.Volume{
					Name:      volume.Name,
					PvcName:   &pvcName,
					MountPath: volume.AppPath,
					Init:      volume.Init,
				})
			}

			var envVars []models.EnvVar
			for _, env := range mainApp.Envs {
				if env.Name == "PORT" {
					continue
				}

				envVars = append(envVars, models.EnvVar{
					Name:  env.Name,
					Value: env.Value,
				})
			}
			envVars = append(envVars, models.EnvVar{
				Name:  "PORT",
				Value: fmt.Sprintf("%d", kg.d.deployment.GetMainApp().InternalPort),
			})

			k8sDeployment.Volumes = volumes
			k8sDeployment.EnvVars = envVars
			k8sDeployment.Image = mainApp.Image
			k8sDeployment.InitCommands = mainApp.InitCommands

			res = append(res, *k8sDeployment)
		} else {
			var imagePullSecrets []string
			if kg.d.deployment.Type == deployment.TypeCustom {
				imagePullSecrets = []string{constants.ImagePullSecretSuffix(kg.d.deployment.Name)}
			}

			mainApp := kg.d.deployment.GetMainApp()

			k8sEnvs := make([]models.EnvVar, len(mainApp.Envs))
			for i, env := range mainApp.Envs {
				if env.Name == "PORT" {
					continue
				}

				k8sEnvs[i] = models.EnvVar{
					Name:  env.Name,
					Value: env.Value,
				}
			}

			k8sEnvs = append(k8sEnvs, models.EnvVar{
				Name:  "PORT",
				Value: fmt.Sprintf("%d", mainApp.InternalPort),
			})

			defaultLimits := models.Limits{
				CPU:    conf.Env.Deployment.Resources.Limits.CPU,
				Memory: conf.Env.Deployment.Resources.Limits.Memory,
			}

			defaultRequests := models.Requests{
				CPU:    conf.Env.Deployment.Resources.Requests.CPU,
				Memory: conf.Env.Deployment.Resources.Requests.Memory,
			}

			k8sVolumes := make([]models.Volume, len(mainApp.Volumes))
			for i, volume := range mainApp.Volumes {
				pvcName := fmt.Sprintf("%s-%s", kg.d.deployment.Name, volume.Name)
				k8sVolumes[i] = models.Volume{
					Name:      volume.Name,
					PvcName:   &pvcName,
					MountPath: volume.AppPath,
					Init:      volume.Init,
				}
			}

			res = append(res, models.DeploymentPublic{
				Name:      kg.d.deployment.Name,
				Namespace: kg.namespace,
				Image:     mainApp.Image,
				EnvVars:   k8sEnvs,
				Resources: models.Resources{
					Limits:   defaultLimits,
					Requests: defaultRequests,
				},
				Volumes:          k8sVolumes,
				ImagePullSecrets: imagePullSecrets,
			})
		}

		return res
	}

	if kg.v.vm != nil {
		ports := kg.v.vm.Ports

		for mapName, k8sDeployment := range kg.v.vm.Subsystems.K8s.GetDeploymentMap() {
			idx := slices.IndexFunc(ports, func(p vm.Port) bool {
				if p.HttpProxy == nil {
					return false
				}

				return getVmProxyDeploymentName(kg.v.vm, p.HttpProxy.Name) == mapName
			})

			if idx != -1 {
				csPort := kg.v.vm.Subsystems.CS.GetPortForwardingRule(ports[idx].Name)
				if csPort == nil {
					continue
				}

				envVars := []models.EnvVar{
					{
						Name:  "PORT",
						Value: "8080",
					},
					{
						Name:  "VM_PORT",
						Value: fmt.Sprintf("%d", csPort.PublicPort),
					},
					{
						Name:  "URL",
						Value: getVmProxyExternalURL(ports[idx].HttpProxy.Name, kg.v.deploymentZone),
					},
					{
						Name:  "VM_URL",
						Value: kg.v.vmZone.ParentDomain,
					},
				}

				k8sDeployment.EnvVars = envVars

				res = append(res, k8sDeployment)
			}
		}

		for _, port := range ports {
			if port.HttpProxy == nil {
				continue
			}

			csPort := kg.v.vm.Subsystems.CS.GetPortForwardingRule(port.Name)
			if csPort == nil {
				continue
			}

			if _, ok := kg.v.vm.Subsystems.K8s.GetDeploymentMap()[getVmProxyDeploymentName(kg.v.vm, port.HttpProxy.Name)]; !ok {
				envVars := []models.EnvVar{
					{
						Name:  "PORT",
						Value: "8080",
					},
					{
						Name:  "VM_PORT",
						Value: fmt.Sprintf("%d", csPort.PublicPort),
					},
					{
						Name:  "URL",
						Value: getVmProxyExternalURL(port.HttpProxy.Name, kg.v.deploymentZone),
					},
					{
						Name:  "VM_URL",
						Value: kg.v.vmZone.ParentDomain,
					},
				}

				res = append(res, models.DeploymentPublic{
					Name:      getVmProxyDeploymentName(kg.v.vm, port.HttpProxy.Name),
					Namespace: kg.namespace,
					Image:     conf.Env.Registry.VmHttpProxyImage,
					EnvVars:   envVars,
					Resources: models.Resources{
						Limits: models.Limits{
							CPU:    conf.Env.Deployment.Resources.Limits.CPU,
							Memory: conf.Env.Deployment.Resources.Limits.Memory,
						},
						Requests: models.Requests{
							CPU:    conf.Env.Deployment.Resources.Requests.CPU,
							Memory: conf.Env.Deployment.Resources.Requests.Memory,
						},
					},
				})
			}
		}

		return res
	}

	if kg.s.storageManager != nil {
		// filebrowser
		if filebrowser := kg.s.storageManager.Subsystems.K8s.GetDeployment(constants.StorageManagerAppName); service.Created(filebrowser) {
			res = append(res, *filebrowser)
		} else {
			initVolumes, volumes := getStorageManagerVolumes(kg.s.storageManager.OwnerID)
			allVolumes := append(initVolumes, volumes...)

			k8sVolumes := make([]models.Volume, len(allVolumes))
			for i, volume := range allVolumes {
				pvcName := getStorageManagerPvcName(volume.Name)
				k8sVolumes[i] = models.Volume{
					Name:      getStorageManagerPvName(kg.s.storageManager.OwnerID, volume.Name),
					PvcName:   &pvcName,
					MountPath: volume.AppPath,
					Init:      volume.Init,
				}
			}

			defaultLimits := models.Limits{
				CPU:    conf.Env.Deployment.Resources.Limits.CPU,
				Memory: conf.Env.Deployment.Resources.Limits.Memory,
			}

			defaultRequests := models.Requests{
				CPU:    conf.Env.Deployment.Resources.Requests.CPU,
				Memory: conf.Env.Deployment.Resources.Requests.Memory,
			}

			args := []string{
				"--noauth",
				"--root=/deploy",
				"--database=/data/database.db",
				"--port=80",
			}

			res = append(res, models.DeploymentPublic{
				Name:      constants.StorageManagerAppName,
				Namespace: kg.namespace,
				Image:     "filebrowser/filebrowser",
				Resources: models.Resources{
					Limits:   defaultLimits,
					Requests: defaultRequests,
				},
				Args:    args,
				Volumes: k8sVolumes,
			})
		}

		// oauth2-proxy
		if oauthProxy := kg.s.storageManager.Subsystems.K8s.GetDeployment(constants.StorageManagerAppNameAuth); service.Created(oauthProxy) {
			res = append(res, *oauthProxy)
		} else {
			defaultLimits := models.Limits{
				CPU:    conf.Env.Deployment.Resources.Limits.CPU,
				Memory: conf.Env.Deployment.Resources.Limits.Memory,
			}

			defaultRequests := models.Requests{
				CPU:    conf.Env.Deployment.Resources.Requests.CPU,
				Memory: conf.Env.Deployment.Resources.Requests.Memory,
			}

			user, err := userModel.New().GetByID(kg.s.storageManager.OwnerID)
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to get user by id when creating oauth proxy deployment public. details: %w", err))
				return nil
			}

			volumes := []models.Volume{
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
			redirectURL := fmt.Sprintf("https://%s.%s/oauth2/callback", kg.s.storageManager.OwnerID, kg.s.zone.Storage.ParentDomain)
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
				"--code-challenge-method=S256",
				"--oidc-groups-claim=groups",
				"--allowed-group=" + conf.Env.Keycloak.AdminGroup,
				"--authenticated-emails-file=/mnt/authenticated-emails-list",
			}

			initContainers := []models.InitContainer{
				{
					Name:    "oauth-proxy-config-init",
					Image:   "busybox",
					Command: []string{"sh", "-c", fmt.Sprintf("mkdir -p /mnt/config && echo %s > /mnt/config/authenticated-emails-list", user.Email)},
					Args:    nil,
				},
			}

			res = append(res, models.DeploymentPublic{
				Name:      constants.StorageManagerAppNameAuth,
				Namespace: kg.namespace,
				Image:     "quay.io/oauth2-proxy/oauth2-proxy:latest",
				EnvVars:   nil,
				Resources: models.Resources{
					Limits:   defaultLimits,
					Requests: defaultRequests,
				},
				Args:           args,
				InitContainers: initContainers,
				Volumes:        volumes,
			})
		}

		return res
	}

	return nil
}

func (kg *K8sGenerator) Services() []models.ServicePublic {
	var res []models.ServicePublic
	if kg.d.deployment != nil {
		if k8sService := kg.d.deployment.Subsystems.K8s.GetService(kg.d.deployment.Name); service.Created(k8sService) {
			res = append(res, *k8sService)
		} else {
			mainApp := kg.d.deployment.GetMainApp()
			res = append(res, models.ServicePublic{
				Name:       kg.d.deployment.Name,
				Namespace:  kg.namespace,
				Port:       mainApp.InternalPort,
				TargetPort: mainApp.InternalPort,
			})
		}
		return res
	}

	if kg.v.vm != nil {
		ports := kg.v.vm.Ports

		for mapName, svc := range kg.v.vm.Subsystems.K8s.GetServiceMap() {
			if slices.IndexFunc(ports, func(p vm.Port) bool {
				if p.HttpProxy == nil {
					return false
				}

				return getVmProxyServiceName(kg.v.vm, p.HttpProxy.Name) == mapName
			}) != -1 {
				res = append(res, svc)
			}
		}

		for _, port := range ports {
			if port.HttpProxy == nil {
				continue
			}

			if _, ok := kg.v.vm.Subsystems.K8s.GetServiceMap()[getVmProxyServiceName(kg.v.vm, port.HttpProxy.Name)]; !ok {
				res = append(res, models.ServicePublic{
					Name:       getVmProxyServiceName(kg.v.vm, port.HttpProxy.Name),
					Namespace:  kg.namespace,
					Port:       8080,
					TargetPort: 8080,
				})
			}
		}

		return res
	}

	if kg.s.storageManager != nil {
		// filebrowser
		if filebrowser := kg.s.storageManager.Subsystems.K8s.GetService(constants.StorageManagerAppName); service.Created(filebrowser) {
			res = append(res, *filebrowser)
		} else {
			res = append(res, models.ServicePublic{
				Name:       constants.StorageManagerAppName,
				Namespace:  kg.namespace,
				Port:       80,
				TargetPort: 80,
			})
		}

		// oauth2-proxy
		if oauthProxy := kg.s.storageManager.Subsystems.K8s.GetService(constants.StorageManagerAppNameAuth); service.Created(oauthProxy) {
			res = append(res, *oauthProxy)
		} else {
			res = append(res, models.ServicePublic{
				Name:       constants.StorageManagerAppNameAuth,
				Namespace:  kg.namespace,
				Port:       4180,
				TargetPort: 4180,
			})
		}
		return res
	}

	return nil
}

func (kg *K8sGenerator) Ingresses() []models.IngressPublic {
	var res []models.IngressPublic
	if kg.d.deployment != nil {
		if !kg.d.deployment.GetMainApp().Private {

			if k8sIngress := kg.d.deployment.Subsystems.K8s.GetIngress(kg.d.deployment.Name); service.Created(k8sIngress) {
				k8sIngress.Hosts = []string{getExternalFQDN(kg.d.deployment.Name, kg.d.zone)}

				res = append(res, *k8sIngress)
			} else {
				tlsSecret := constants.WildcardCertSecretName

				res = append(res, models.IngressPublic{
					Name:         kg.d.deployment.Name,
					Namespace:    kg.namespace,
					ServiceName:  kg.d.deployment.Name,
					ServicePort:  kg.d.deployment.GetMainApp().InternalPort,
					IngressClass: conf.Env.Deployment.IngressClass,
					Hosts:        []string{getExternalFQDN(kg.d.deployment.Name, kg.d.zone)},
					TlsSecret:    &tlsSecret,
				})
			}

			if customK8sIngress := kg.d.deployment.Subsystems.K8s.GetIngress(constants.CustomDomainSuffix(kg.d.deployment.Name)); service.Created(customK8sIngress) {
				customK8sIngress.Hosts = []string{*kg.d.deployment.GetMainApp().CustomDomain}
				res = append(res, *customK8sIngress)
			} else {
				if kg.d.deployment.GetMainApp().CustomDomain != nil {
					res = append(res, models.IngressPublic{
						Name:         fmt.Sprintf(constants.CustomDomainSuffix(kg.d.deployment.Name)),
						Namespace:    kg.namespace,
						ServiceName:  kg.d.deployment.Name,
						ServicePort:  kg.d.deployment.GetMainApp().InternalPort,
						IngressClass: conf.Env.Deployment.IngressClass,
						Hosts:        []string{*kg.d.deployment.GetMainApp().CustomDomain},
						CustomCert: &models.CustomCert{
							ClusterIssuer: "letsencrypt-prod-deploy-http",
							CommonName:    *kg.d.deployment.GetMainApp().CustomDomain,
						},
					})
				}
			}
		}
		return res
	}

	if kg.v.vm != nil {
		ports := kg.v.vm.Ports

		for mapName, ingress := range kg.v.vm.Subsystems.K8s.GetIngressMap() {
			idx := slices.IndexFunc(ports, func(p vm.Port) bool {
				if p.HttpProxy == nil {
					return false
				}

				return getVmProxyIngressName(kg.v.vm, p.HttpProxy.Name) == mapName ||
					(getVmProxyCustomDomainIngressName(kg.v.vm, p.HttpProxy.Name) == mapName && p.HttpProxy.CustomDomain != nil)
			})

			if idx != -1 {
				if getVmProxyCustomDomainIngressName(kg.v.vm, ports[idx].HttpProxy.Name) == mapName {
					if ports[idx].HttpProxy.CustomDomain != nil {
						ingress.Hosts = []string{*ports[idx].HttpProxy.CustomDomain}
					}
				} else {
					ingress.Hosts = []string{getVmProxyExternalURL(ports[idx].HttpProxy.Name, kg.v.deploymentZone)}
				}

				res = append(res, ingress)
			}
		}

		for _, port := range ports {
			if port.HttpProxy == nil {
				continue
			}

			if _, ok := kg.v.vm.Subsystems.K8s.GetIngressMap()[getVmProxyIngressName(kg.v.vm, port.HttpProxy.Name)]; !ok {
				tlsSecret := constants.WildcardCertSecretName

				res = append(res, models.IngressPublic{
					Name:         getVmProxyIngressName(kg.v.vm, port.HttpProxy.Name),
					Namespace:    kg.namespace,
					ServiceName:  getVmProxyServiceName(kg.v.vm, port.HttpProxy.Name),
					ServicePort:  8080,
					IngressClass: conf.Env.Deployment.IngressClass,
					Hosts:        []string{getVmProxyExternalURL(port.HttpProxy.Name, kg.v.deploymentZone)},
					TlsSecret:    &tlsSecret,
				})
			}

			if port.HttpProxy.CustomDomain != nil {
				if _, ok := kg.v.vm.Subsystems.K8s.GetIngressMap()[getVmProxyCustomDomainIngressName(kg.v.vm, port.HttpProxy.Name)]; !ok {
					res = append(res, models.IngressPublic{
						Name:         getVmProxyCustomDomainIngressName(kg.v.vm, port.HttpProxy.Name),
						Namespace:    kg.namespace,
						ServiceName:  getVmProxyServiceName(kg.v.vm, port.HttpProxy.Name),
						ServicePort:  8080,
						IngressClass: conf.Env.Deployment.IngressClass,
						Hosts:        []string{*port.HttpProxy.CustomDomain},
						CustomCert: &models.CustomCert{
							ClusterIssuer: "letsencrypt-prod-deploy-http",
							CommonName:    *port.HttpProxy.CustomDomain,
						},
					})
				}
			}
		}

		return res
	}

	if kg.s.storageManager != nil {
		if ingress := kg.s.storageManager.Subsystems.K8s.GetIngress(constants.StorageManagerAppName); service.Created(ingress) {
			res = append(res, *ingress)
		} else {
			tlsSecret := constants.WildcardCertSecretName

			res = append(res, models.IngressPublic{
				Name:         constants.StorageManagerAppName,
				Namespace:    kg.namespace,
				ServiceName:  constants.StorageManagerAppName,
				ServicePort:  80,
				IngressClass: conf.Env.Deployment.IngressClass,
				Hosts:        []string{getExternalFQDN(kg.s.storageManager.OwnerID, kg.s.zone)},
				TlsSecret:    &tlsSecret,
			})
		}

		return res
	}

	return nil
}

func (kg *K8sGenerator) PrivateIngress() *models.IngressPublic {
	return &models.IngressPublic{
		Placeholder: true,
	}
}

func (kg *K8sGenerator) PVs() []models.PvPublic {
	var res []models.PvPublic

	if kg.d.deployment != nil {
		volumes := kg.d.deployment.GetMainApp().Volumes

		for mapName, pv := range kg.d.deployment.Subsystems.K8s.GetPvMap() {
			if slices.IndexFunc(volumes, func(v deployment.Volume) bool { return v.Name == mapName }) != -1 {
				res = append(res, pv)
			}
		}

		for _, volume := range kg.d.deployment.GetMainApp().Volumes {
			if _, ok := kg.d.deployment.Subsystems.K8s.GetPvMap()[volume.Name]; !ok {
				res = append(res, models.PvPublic{
					Name:      getDeploymentPvName(kg.d.deployment, volume.Name),
					Capacity:  conf.Env.Deployment.Resources.Limits.Storage,
					NfsServer: kg.d.zone.Storage.NfsServer,
					NfsPath:   path.Join(kg.d.zone.Storage.NfsParentPath, kg.d.deployment.OwnerID, "user"),
				})
			}
		}

		return res
	}

	if kg.s.storageManager != nil {
		initVolumes, volumes := getStorageManagerVolumes(kg.s.storageManager.OwnerID)
		allVolumes := append(initVolumes, volumes...)

		for mapName, pv := range kg.s.storageManager.Subsystems.K8s.GetPvMap() {
			if slices.IndexFunc(allVolumes, func(v storageManagerModel.Volume) bool { return v.Name == mapName }) != -1 {
				res = append(res, pv)
			}
		}

		for _, volume := range allVolumes {
			if _, ok := kg.s.storageManager.Subsystems.K8s.GetPvMap()[volume.Name]; !ok {
				res = append(res, models.PvPublic{
					Name:      getStorageManagerPvName(kg.s.storageManager.OwnerID, volume.Name),
					Capacity:  conf.Env.Deployment.Resources.Limits.Storage,
					NfsServer: kg.s.zone.Storage.NfsServer,
					NfsPath:   path.Join(kg.s.zone.Storage.NfsParentPath, volume.ServerPath),
				})
			}
		}
	}

	return res
}

func (kg *K8sGenerator) PVCs() []models.PvcPublic {
	var res []models.PvcPublic

	if kg.d.deployment != nil {
		volumes := kg.d.deployment.GetMainApp().Volumes

		for mapName, pvc := range kg.d.deployment.Subsystems.K8s.GetPvcMap() {
			if slices.IndexFunc(volumes, func(v deployment.Volume) bool { return v.Name == mapName }) != -1 {
				res = append(res, pvc)
			}
		}

		for _, volume := range kg.d.deployment.GetMainApp().Volumes {
			if _, ok := kg.d.deployment.Subsystems.K8s.GetPvcMap()[volume.Name]; !ok {
				res = append(res, models.PvcPublic{
					Name:      getDeploymentPvcName(kg.d.deployment, volume.Name),
					Namespace: kg.namespace,
					Capacity:  conf.Env.Deployment.Resources.Limits.Storage,
					PvName:    getDeploymentPvName(kg.d.deployment, volume.Name),
				})
			}
		}

		return res
	}

	if kg.s.storageManager != nil {
		initVolumes, volumes := getStorageManagerVolumes(kg.s.storageManager.OwnerID)
		allVolumes := append(initVolumes, volumes...)

		for mapName, pvc := range kg.s.storageManager.Subsystems.K8s.GetPvcMap() {
			if slices.IndexFunc(allVolumes, func(v storageManagerModel.Volume) bool { return v.Name == mapName }) != -1 {
				res = append(res, pvc)
			}
		}

		for _, volume := range allVolumes {
			if _, ok := kg.s.storageManager.Subsystems.K8s.GetPvcMap()[volume.Name]; !ok {
				res = append(res, models.PvcPublic{
					Name:      getStorageManagerPvcName(volume.Name),
					Namespace: kg.namespace,
					Capacity:  conf.Env.Deployment.Resources.Limits.Storage,
					PvName:    getStorageManagerPvName(kg.s.storageManager.OwnerID, volume.Name),
				})
			}
		}

		return res
	}

	return res
}

func (kg *K8sGenerator) Secrets() []models.SecretPublic {
	var res []models.SecretPublic

	if kg.d.deployment != nil {
		if kg.d.deployment.Type == deployment.TypeCustom {
			if secret := kg.d.deployment.Subsystems.K8s.GetSecret(constants.ImagePullSecretSuffix(kg.d.deployment.Name)); service.Created(secret) {
				res = append(res, *secret)
			} else if kg.d.deployment.Subsystems.Harbor.Robot.Created() && kg.d.deployment.Type == deployment.TypeCustom {
				registry := conf.Env.Registry.URL
				username := kg.d.deployment.Subsystems.Harbor.Robot.HarborName
				password := kg.d.deployment.Subsystems.Harbor.Robot.Secret

				res = append(res, models.SecretPublic{
					Name:      constants.ImagePullSecretSuffix(kg.d.deployment.Name),
					Namespace: kg.namespace,
					Type:      string(v1.SecretTypeDockerConfigJson),
					Data: map[string][]byte{
						v1.DockerConfigJsonKey: encodeDockerConfig(registry, username, password),
					},
				})
			}
		}

		// wildcard certificate
		if secret := kg.d.deployment.Subsystems.K8s.GetSecret(constants.WildcardCertSecretName); service.Created(secret) {
			res = append(res, *secret)
		} else {
			// swap namespaces temporarily
			kg.client.Namespace = conf.Env.Deployment.WildcardCertSecretNamespace
			defer func() { kg.client.Namespace = kg.namespace }()

			copyFrom, err := kg.client.ReadSecret(conf.Env.Deployment.WildcardCertSecretId)
			if err != nil || copyFrom == nil {
				utils.PrettyPrintError(fmt.Errorf("failed to read secret %s/%s. details: %w", conf.Env.Deployment.WildcardCertSecretNamespace, conf.Env.Deployment.WildcardCertSecretId, err))
			} else {
				res = append(res, models.SecretPublic{
					Name:      constants.WildcardCertSecretName,
					Namespace: kg.namespace,
					Type:      string(v1.SecretTypeOpaque),
					Data:      copyFrom.Data,
				})
			}
		}

		return res
	}

	if kg.v.vm != nil {
		// wildcard certificate
		if secret := kg.v.vm.Subsystems.K8s.GetSecret(constants.WildcardCertSecretName); service.Created(secret) {
			res = append(res, *secret)
		} else {
			// swap namespaces temporarily
			kg.client.Namespace = conf.Env.Deployment.WildcardCertSecretNamespace
			defer func() { kg.client.Namespace = kg.namespace }()

			copyFrom, err := kg.client.ReadSecret(conf.Env.Deployment.WildcardCertSecretId)
			if err != nil || copyFrom == nil {
				utils.PrettyPrintError(fmt.Errorf("failed to read secret %s/%s. details: %w", conf.Env.Deployment.WildcardCertSecretNamespace, conf.Env.Deployment.WildcardCertSecretId, err))
			} else {
				res = append(res, models.SecretPublic{
					Name:      constants.WildcardCertSecretName,
					Namespace: kg.namespace,
					Type:      string(v1.SecretTypeOpaque),
					Data:      copyFrom.Data,
				})
			}
		}

		return res
	}

	if kg.s.storageManager != nil {
		// wildcard certificate
		if secret := kg.s.storageManager.Subsystems.K8s.GetSecret(constants.WildcardCertSecretName); service.Created(secret) {
			res = append(res, *secret)
		} else {
			// swap namespaces temporarily
			kg.client.Namespace = conf.Env.Deployment.WildcardCertSecretNamespace
			defer func() { kg.client.Namespace = kg.namespace }()

			copyFrom, err := kg.client.ReadSecret(conf.Env.Deployment.WildcardCertSecretId)
			if err != nil || copyFrom == nil {
				utils.PrettyPrintError(fmt.Errorf("failed to read secret %s/%s. details: %w", conf.Env.Deployment.WildcardCertSecretNamespace, conf.Env.Deployment.WildcardCertSecretId, err))
			} else {
				res = append(res, models.SecretPublic{
					Name:      constants.WildcardCertSecretName,
					Namespace: kg.namespace,
					Type:      string(v1.SecretTypeOpaque),
					Data:      copyFrom.Data,
				})
			}
		}

		return res
	}

	return nil
}

func (kg *K8sGenerator) Jobs() []models.JobPublic {
	var res []models.JobPublic

	if kg.s.storageManager != nil {
		jobs := getStorageManagerJobs(kg.s.storageManager.OwnerID)

		for mapName, job := range kg.s.storageManager.Subsystems.K8s.GetJobMap() {
			if slices.IndexFunc(jobs, func(j storageManagerModel.Job) bool { return j.Name == mapName }) != -1 {
				res = append(res, job)
			}
		}

		for _, job := range jobs {
			if _, ok := kg.s.storageManager.Subsystems.K8s.GetJobMap()[job.Name]; !ok {
				res = append(res, models.JobPublic{
					Name:      job.Name,
					Namespace: kg.namespace,
					Image:     job.Image,
					Command:   job.Command,
					Args:      job.Args,
				})
			}
		}

		return res
	}

	return nil
}

func getExternalFQDN(name string, zone *enviroment.DeploymentZone) string {
	return fmt.Sprintf("%s.%s", name, zone.ParentDomain)
}

func encodeDockerConfig(registry, username, password string) []byte {
	dockerConfig := map[string]interface{}{
		"auths": map[string]interface{}{
			registry: map[string]interface{}{
				"username": username,
				"password": password,
				"auth":     base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password))),
			},
		},
	}

	jsonData, err := json.Marshal(dockerConfig)
	if err != nil {
		jsonData = make([]byte, 0)
	}

	return jsonData
}

func getDeploymentPvName(deployment *deployment.Deployment, volumeName string) string {
	return fmt.Sprintf("%s-%s", deployment.Name, volumeName)
}

func getDeploymentPvcName(deployment *deployment.Deployment, volumeName string) string {
	return fmt.Sprintf("%s-%s", deployment.Name, volumeName)
}

func getVmProxyDeploymentName(vm *vm.VM, portName string) string {
	return fmt.Sprintf("%s-%s", vm.Name, portName)
}

func getVmProxyServiceName(vm *vm.VM, portName string) string {
	return fmt.Sprintf("%s-%s", vm.Name, portName)
}

func getVmProxyIngressName(vm *vm.VM, portName string) string {
	return fmt.Sprintf("%s-%s", vm.Name, portName)
}

func getVmProxyCustomDomainIngressName(vm *vm.VM, portName string) string {
	return fmt.Sprintf("%s-%s-custom-domain", vm.Name, portName)
}

func getVmProxyExternalURL(portName string, zone *enviroment.DeploymentZone) string {
	return fmt.Sprintf("%s.%s", portName, zone.ParentDomainVM)
}

func getStorageManagerPvcName(volumeName string) string {
	return fmt.Sprintf("%s-%s", constants.StorageManagerAppName, volumeName)
}

func getStorageManagerPvName(ownerID, volumeName string) string {
	return fmt.Sprintf("%s-%s", volumeName, ownerID)
}

func getStorageManagerVolumes(ownerID string) ([]storageManagerModel.Volume, []storageManagerModel.Volume) {
	initVolumes := []storageManagerModel.Volume{
		{
			Name:       "init",
			Init:       false,
			AppPath:    "/exports",
			ServerPath: "",
		},
	}

	volumes := []storageManagerModel.Volume{
		{
			Name:       "data",
			Init:       false,
			AppPath:    "/data",
			ServerPath: path.Join(ownerID, "data"),
		},
		{
			Name:       "user",
			Init:       false,
			AppPath:    "/deploy",
			ServerPath: path.Join(ownerID, "user"),
		},
	}

	return initVolumes, volumes
}

func getStorageManagerJobs(userID string) []storageManagerModel.Job {
	return []storageManagerModel.Job{
		{
			Name:    "init",
			Image:   "busybox",
			Command: []string{"/bin/mkdir"},
			Args: []string{
				"-p",
				path.Join("/exports", userID, "data"),
				path.Join("/exports", userID, "user"),
			},
		},
	}
}

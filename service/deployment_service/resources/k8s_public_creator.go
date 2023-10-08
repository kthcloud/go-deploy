package resources

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go-deploy/models/sys/deployment"
	storageManagerModel "go-deploy/models/sys/deployment/storage_manager"
	"go-deploy/models/sys/enviroment"
	userModel "go-deploy/models/sys/user"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service"
	"go-deploy/service/deployment_service/base"
	"go-deploy/utils"
	"go-deploy/utils/subsystemutils"
	"golang.org/x/exp/slices"
	v1 "k8s.io/api/core/v1"
	"path"
)

type K8sGenerator struct {
	*PublicGeneratorType
	namespace string
}

func (kg *K8sGenerator) MainNamespace() *models.NamespacePublic {
	return &models.NamespacePublic{
		Name: getNamespaceName(kg.deployment.OwnerID),
	}
}

func (kg *K8sGenerator) StorageManagerNamespace() *models.NamespacePublic {
	return &models.NamespacePublic{
		Name: getStorageManagerNamespaceName(kg.deployment.OwnerID),
	}
}

func (kg *K8sGenerator) Deployments() []models.DeploymentPublic {
	var res []models.DeploymentPublic

	if kg.deployment != nil {
		if k8sDeployment := kg.deployment.Subsystems.K8s.GetDeployment(base.AppName); service.Created(k8sDeployment) {
			if kg.updateParams != nil {
				if kg.updateParams.Volumes != nil {
					var volumes []models.Volume
					for _, volume := range *kg.updateParams.Volumes {
						volumes = append(volumes, models.Volume{
							Name:      volume.Name,
							PvcName:   nil,
							MountPath: volume.AppPath,
							Init:      volume.Init,
						})
					}
					k8sDeployment.Volumes = volumes
				}

				if kg.updateParams.Envs != nil {
					var envVars []models.EnvVar
					for _, env := range *kg.updateParams.Envs {
						envVars = append(envVars, models.EnvVar{
							Name:  env.Name,
							Value: env.Value,
						})
					}
					k8sDeployment.EnvVars = envVars
				}

				if kg.updateParams.Image != nil {
					k8sDeployment.Image = *kg.updateParams.Image
				}

				if kg.updateParams.InitCommands != nil {
					k8sDeployment.InitCommands = *kg.updateParams.InitCommands
				}
			}

			res = append(res, *k8sDeployment)
		} else {
			var imagePullSecrets []string
			if kg.deployment.Type == deployment.TypeCustom {
				imagePullSecrets = []string{base.AppNameImagePullSecret}
			}

			mainApp := kg.deployment.GetMainApp()

			k8sEnvs := make([]models.EnvVar, len(mainApp.Envs))
			for i, env := range mainApp.Envs {
				k8sEnvs[i] = models.EnvVar{
					Name:  env.Name,
					Value: env.Value,
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

			k8sVolumes := make([]models.Volume, len(mainApp.Volumes))
			for i, volume := range mainApp.Volumes {
				pvcName := fmt.Sprintf("%s-%s", base.AppName, volume.Name)
				k8sVolumes[i] = models.Volume{
					Name:      volume.Name,
					PvcName:   &pvcName,
					MountPath: volume.AppPath,
					Init:      volume.Init,
				}
			}

			res = append(res, models.DeploymentPublic{
				Name:      base.AppName,
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

	if kg.storageManager != nil {
		// filebrowser
		if filebrowser := kg.deployment.Subsystems.K8s.GetDeployment(base.StorageManagerAppName); service.Created(filebrowser) {
			res = append(res, *filebrowser)
		} else {
			_, volumes := getStorageManagerVolumes(kg.storageManager.OwnerID)

			k8sVolumes := make([]models.Volume, len(volumes))
			for i, volume := range volumes {
				pvcName := volume.Name
				k8sVolumes[i] = models.Volume{
					Name:      volume.Name,
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
				Name:      base.StorageManagerAppName,
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
		if oauthProxy := kg.deployment.Subsystems.K8s.GetDeployment(base.StorageManagerAppNameAuth); service.Created(oauthProxy) {
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

			user, err := userModel.New().GetByID(kg.storageManager.OwnerID)
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
			redirectURL := fmt.Sprintf("https://%s.%s/oauth2/callback", kg.storageManager.OwnerID, kg.zone.Storage.ParentDomain)
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
				Name:      base.StorageManagerAppNameAuth,
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
	if kg.deployment != nil {
		if k8sService := kg.deployment.Subsystems.K8s.GetService(base.AppName); service.Created(k8sService) {
			res = append(res, *k8sService)
		} else {
			mainApp := kg.deployment.GetMainApp()
			res = append(res, models.ServicePublic{
				Name:       base.AppName,
				Namespace:  kg.namespace,
				Port:       mainApp.InternalPort,
				TargetPort: mainApp.InternalPort,
			})
		}
		return res
	}

	if kg.storageManager != nil {
		// filebrowser
		if filebrowser := kg.deployment.Subsystems.K8s.GetService(base.StorageManagerAppName); service.Created(filebrowser) {
			res = append(res, *filebrowser)
		} else {
			res = append(res, models.ServicePublic{
				Name:       base.StorageManagerAppName,
				Namespace:  kg.namespace,
				Port:       80,
				TargetPort: 80,
			})
		}

		// oauth2-proxy
		if oauthProxy := kg.deployment.Subsystems.K8s.GetService(base.StorageManagerAppNameAuth); service.Created(oauthProxy) {
			res = append(res, *oauthProxy)
		} else {
			res = append(res, models.ServicePublic{
				Name:       base.StorageManagerAppNameAuth,
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
	if kg.deployment != nil {
		if !kg.deployment.GetMainApp().Private {
			res = append(res, models.IngressPublic{
				Name:         kg.deployment.Name,
				Namespace:    kg.namespace,
				ServiceName:  kg.deployment.Name,
				ServicePort:  kg.deployment.GetMainApp().InternalPort,
				IngressClass: conf.Env.Deployment.IngressClass,
				Hosts:        []string{getExternalFQDN(kg.deployment.Name, kg.zone)},
			})

			var customDomain *string
			if kg.createParams != nil && kg.createParams.CustomDomain != nil {
				customDomain = kg.createParams.CustomDomain
			} else {
				customDomain = kg.deployment.GetMainApp().CustomDomain
			}

			if customDomain != nil {
				res = append(res, models.IngressPublic{
					Name:         fmt.Sprintf("%s-%s", kg.deployment.Name, base.AppNameCustomDomain),
					Namespace:    kg.namespace,
					ServiceName:  kg.deployment.Name,
					ServicePort:  kg.deployment.GetMainApp().InternalPort,
					IngressClass: conf.Env.Deployment.IngressClass,
					Hosts:        []string{*kg.createParams.CustomDomain},
					CustomCert: &models.CustomCert{
						ClusterIssuer: "letsencrypt-prod-deploy-http",
						CommonName:    *kg.createParams.CustomDomain,
					},
				})
			}
		}
		return res
	}

	if kg.storageManager != nil {
		if ingress := kg.deployment.Subsystems.K8s.GetIngress(base.StorageManagerAppName); service.Created(ingress) {
			res = append(res, *ingress)
		} else {
			res = append(res, models.IngressPublic{
				Name:         base.StorageManagerAppName,
				Namespace:    kg.namespace,
				ServiceName:  base.StorageManagerAppName,
				ServicePort:  80,
				IngressClass: conf.Env.Deployment.IngressClass,
				Hosts:        []string{getExternalFQDN(kg.storageManager.OwnerID, kg.zone)},
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

	if kg.deployment != nil {
		volumes := kg.deployment.GetMainApp().Volumes

		for mapName, pv := range kg.deployment.Subsystems.K8s.PvMap {
			if slices.IndexFunc(volumes, func(v deployment.Volume) bool { return v.Name == mapName }) != -1 {
				res = append(res, pv)
			}
		}

		for _, volume := range kg.deployment.GetMainApp().Volumes {
			if _, ok := kg.deployment.Subsystems.K8s.PvMap[volume.Name]; !ok {
				res = append(res, models.PvPublic{
					Name:      getPvName(kg.deployment, volume.Name),
					Capacity:  conf.Env.Deployment.Resources.Limits.Storage,
					NfsServer: kg.zone.Storage.NfsServer,
					NfsPath:   path.Join(kg.zone.Storage.NfsParentPath, kg.deployment.OwnerID, "user"),
				})
			}
		}

		return res
	}

	if kg.storageManager != nil {
		initVolumes, volumes := getStorageManagerVolumes(kg.storageManager.OwnerID)
		allVolumes := append(initVolumes, volumes...)

		for mapName, pv := range kg.storageManager.Subsystems.K8s.PvMap {
			if slices.IndexFunc(allVolumes, func(v storageManagerModel.Volume) bool { return v.Name == mapName }) != -1 {
				res = append(res, pv)
			}
		}

		for _, volume := range allVolumes {
			if _, ok := kg.storageManager.Subsystems.K8s.PvMap[volume.Name]; !ok {
				res = append(res, models.PvPublic{
					Name:      getStorageManagerPvName(kg.storageManager.OwnerID, volume.Name),
					Capacity:  conf.Env.Deployment.Resources.Limits.Storage,
					NfsServer: kg.zone.Storage.NfsServer,
					NfsPath:   path.Join(kg.zone.Storage.NfsParentPath, volume.ServerPath),
				})
			}
		}
	}

	return res
}

func (kg *K8sGenerator) PVCs() []models.PvcPublic {
	var res []models.PvcPublic

	if kg.deployment != nil {
		volumes := kg.deployment.GetMainApp().Volumes

		for mapName, pvc := range kg.deployment.Subsystems.K8s.PvcMap {
			if slices.IndexFunc(volumes, func(v deployment.Volume) bool { return v.Name == mapName }) != -1 {
				res = append(res, pvc)
			}
		}

		for _, volume := range kg.deployment.GetMainApp().Volumes {
			if _, ok := kg.deployment.Subsystems.K8s.PvcMap[volume.Name]; !ok {
				res = append(res, models.PvcPublic{
					Name:      getPvcName(kg.deployment, volume.Name),
					Namespace: kg.namespace,
					Capacity:  conf.Env.Deployment.Resources.Limits.Storage,
					PvName:    getPvName(kg.deployment, volume.Name),
				})
			}
		}

		return res
	}

	if kg.storageManager != nil {
		_, volumes := getStorageManagerVolumes(kg.storageManager.OwnerID)

		for mapName, pvc := range kg.storageManager.Subsystems.K8s.PvcMap {
			if slices.IndexFunc(volumes, func(v storageManagerModel.Volume) bool { return v.Name == mapName }) != -1 {
				res = append(res, pvc)
			}
		}

		for _, volume := range volumes {
			if _, ok := kg.storageManager.Subsystems.K8s.PvcMap[volume.Name]; !ok {
				res = append(res, models.PvcPublic{
					Name:      getStorageManagerPvcName(volume.Name),
					Namespace: kg.namespace,
					Capacity:  conf.Env.Deployment.Resources.Limits.Storage,
					PvName:    getStorageManagerPvName(kg.storageManager.OwnerID, volume.Name),
				})
			}
		}

		return res
	}

	return res
}

func (kg *K8sGenerator) Secrets() []models.SecretPublic {
	var res []models.SecretPublic

	if kg.deployment != nil {
		if !kg.deployment.Subsystems.Harbor.Robot.Created() && kg.deployment.Type == deployment.TypeCustom {
			registry := conf.Env.DockerRegistry.URL
			username := kg.deployment.Subsystems.Harbor.Robot.HarborName
			password := kg.deployment.Subsystems.Harbor.Robot.Secret

			res = append(res, models.SecretPublic{
				Name:      kg.deployment.Name,
				Namespace: kg.namespace,
				Type:      string(v1.SecretTypeDockerConfigJson),
				Data: map[string][]byte{
					v1.DockerConfigJsonKey: encodeDockerConfig(registry, username, password),
				},
			})
		}

		return res
	}

	return nil
}

func (kg *K8sGenerator) Jobs() []models.JobPublic {
	var res []models.JobPublic

	if kg.storageManager != nil {
		jobs := getStorageManagerJobs(kg.storageManager.OwnerID)

		for mapName, job := range kg.storageManager.Subsystems.K8s.JobMap {
			if slices.IndexFunc(jobs, func(j storageManagerModel.Job) bool { return j.Name == mapName }) != -1 {
				res = append(res, job)
			}
		}

		for _, job := range jobs {
			if _, ok := kg.storageManager.Subsystems.K8s.JobMap[job.Name]; !ok {
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

func getNamespaceName(userID string) string {
	return subsystemutils.GetPrefixedName(userID)
}

func getStorageManagerNamespaceName(userID string) string {
	return subsystemutils.GetPrefixedName(fmt.Sprintf("%s-%s", base.StorageManagerNamePrefix, userID))
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

func getPvName(deployment *deployment.Deployment, volumeName string) string {
	return fmt.Sprintf("%s-%s", deployment.Name, volumeName)
}

func getPvcName(deployment *deployment.Deployment, volumeName string) string {
	return fmt.Sprintf("%s-%s", deployment.Name, volumeName)
}

func getStorageManagerPvcName(volumeName string) string {
	return fmt.Sprintf("%s-%s", base.StorageManagerAppName, volumeName)
}

func getStorageManagerPvName(ownerID, volumeName string) string {
	return fmt.Sprintf("%s-%s", volumeName, ownerID)
}

func getStorageManagerVolumes(ownerID string) ([]storageManagerModel.Volume, []storageManagerModel.Volume) {
	initVolumes := []storageManagerModel.Volume{
		{
			Name:       fmt.Sprintf("%s-%s", base.StorageManagerAppName, "init"),
			Init:       false,
			AppPath:    "/exports",
			ServerPath: "",
		},
	}

	volumes := []storageManagerModel.Volume{
		{
			Name:       fmt.Sprintf("%s-%s", base.StorageManagerAppName, "data"),
			Init:       false,
			AppPath:    "/data",
			ServerPath: path.Join(ownerID, "data"),
		},
		{
			Name:       fmt.Sprintf("%s-%s", base.StorageManagerAppName, "user"),
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

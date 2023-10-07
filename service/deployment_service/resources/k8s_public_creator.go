package resources

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go-deploy/models/sys/deployment"
	"go-deploy/models/sys/enviroment"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service/deployment_service/base"
	v1 "k8s.io/api/core/v1"
	"log"
	"path"
	"strconv"
	"time"
)

type K8sGenerator struct {
	*PublicGeneratorType
	namespace string
}

func (kg *K8sGenerator) Namespace() *models.NamespacePublic {
	return &models.NamespacePublic{
		Name: kg.namespace,
	}
}

func (kg *K8sGenerator) MainDeployment() *models.DeploymentPublic {
	var pullSecrets []string
	if kg.deployment.Type == deployment.TypeCustom {
		pullSecrets = []string{base.AppNameImagePullSecret}
	}

	mainApp := kg.deployment.GetMainApp()

	k8sEnvs := []models.EnvVar{
		{Name: "PORT", Value: strconv.Itoa(mainApp.InternalPort)},
	}

	defaultLimits := models.Limits{
		CPU:    conf.Env.Deployment.Resources.Limits.CPU,
		Memory: conf.Env.Deployment.Resources.Limits.Memory,
	}

	defaultRequests := models.Requests{
		CPU:    conf.Env.Deployment.Resources.Requests.CPU,
		Memory: conf.Env.Deployment.Resources.Requests.Memory,
	}

	res := kg.deployment.Subsystems.K8s.GetDeployment(base.AppName)
	if res == nil {
		res = &models.DeploymentPublic{
			ID:               "",
			Namespace:        kg.namespace,
			ImagePullSecrets: pullSecrets,
			EnvVars:          k8sEnvs,
			Resources: models.Resources{
				Limits:   defaultLimits,
				Requests: defaultRequests,
			},
		}
	}

	var volumes []deployment.Volume

	if kg.createParams != nil {
		volumes = kg.createParams.Volumes

		res.Name = kg.createParams.Name
		res.Image = kg.createParams.Image
	} else if kg.updateParams != nil {
		volumes = kg.createParams.Volumes

		if kg.updateParams.Envs != nil {
			for _, env := range *kg.updateParams.Envs {
				res.EnvVars = append(res.EnvVars, models.EnvVar{
					Name:  env.Name,
					Value: env.Value,
				})
			}

			if kg.updateParams.Image != nil {
				res.Image = *kg.updateParams.Image
			}
		}

		if kg.updateParams.Image != nil {
			res.Image = *kg.updateParams.Image
		}

		if kg.updateParams.InitCommands != nil {
			res.InitCommands = *kg.updateParams.InitCommands
		}
	} else {
		volumes = kg.deployment.GetMainApp().Volumes
	}

	k8sVolumes := make([]models.Volume, len(volumes))
	for i, volume := range volumes {
		pvcName := fmt.Sprintf("%s-%s", res.Name, volume.Name)
		k8sVolumes[i] = models.Volume{
			Name:      volume.Name,
			PvcName:   &pvcName,
			MountPath: volume.AppPath,
			Init:      volume.Init,
		}
	}
	res.Volumes = k8sVolumes

	return res
}

func (kg *K8sGenerator) MainService() *models.ServicePublic {
	mainApp := kg.deployment.GetMainApp()

	return &models.ServicePublic{
		Name:       base.AppName,
		Namespace:  kg.namespace,
		Port:       mainApp.InternalPort,
		TargetPort: mainApp.InternalPort,
	}
}

func (kg *K8sGenerator) MainIngress() *models.IngressPublic {
	return &models.IngressPublic{
		Name:         kg.deployment.Name,
		Namespace:    kg.namespace,
		ServiceName:  kg.deployment.Name,
		ServicePort:  kg.deployment.GetMainApp().InternalPort,
		IngressClass: conf.Env.Deployment.IngressClass,
		Hosts:        []string{getExternalFQDN(kg.deployment.Name, kg.zone)},
	}
}

func (kg *K8sGenerator) CustomDomainIngress() *models.IngressPublic {
	return &models.IngressPublic{
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
	}
}

func (kg *K8sGenerator) PrivateIngress() *models.IngressPublic {
	return &models.IngressPublic{
		Placeholder: true,
	}
}

func (kg *K8sGenerator) PVs() []models.PvPublic {
	var res []models.PvPublic
	if kg.createParams != nil {
		res = make([]models.PvPublic, len(kg.createParams.Volumes))

		for i, volume := range kg.createParams.Volumes {
			res[i] = models.PvPublic{
				Name:      getVolumeName(kg.deployment, volume.Name),
				Capacity:  conf.Env.Deployment.Resources.Limits.Storage,
				NfsServer: kg.zone.Storage.NfsServer,
				NfsPath:   path.Join(kg.zone.Storage.NfsParentPath, kg.deployment.OwnerID, "user"),
			}
		}
	} else if kg.updateParams != nil {
		// TODO
	}

	return res
}

func (kg *K8sGenerator) PVCs() []models.PvcPublic {
	var res []models.PvcPublic
	if kg.createParams != nil {
		res = make([]models.PvcPublic, len(kg.createParams.Volumes))

		for i, volume := range kg.createParams.Volumes {
			res[i] = models.PvcPublic{
				Name:      getVolumeName(kg.deployment, volume.Name),
				Namespace: kg.namespace,
				Capacity:  conf.Env.Deployment.Resources.Limits.Storage,
				PvName:    getVolumeName(kg.deployment, volume.Name),
				CreatedAt: time.Time{},
			}
		}
	} else if kg.updateParams != nil {
		// TODO
	}

	return res
}

func (kg *K8sGenerator) ImagePullSecret() *models.SecretPublic {
	if kg.deployment.Type != deployment.TypeCustom {
		return nil
	}

	if !kg.deployment.Subsystems.Harbor.Robot.Created() {
		log.Println("harbor robot not for deployment", kg.deployment.ID, "found when creating k8s image pull secret")
		return nil
	}

	registry := conf.Env.DockerRegistry.URL
	username := kg.deployment.Subsystems.Harbor.Robot.HarborName
	password := kg.deployment.Subsystems.Harbor.Robot.Secret

	return &models.SecretPublic{
		Name:      kg.deployment.Name,
		Namespace: kg.namespace,
		Type:      string(v1.SecretTypeDockerConfigJson),
		Data: map[string][]byte{
			v1.DockerConfigJsonKey: encodeDockerConfig(registry, username, password),
		},
	}
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

func getVolumeName(deployment *deployment.Deployment, volumeName string) string {
	return fmt.Sprintf("%s-%s", deployment.Name, volumeName)
}

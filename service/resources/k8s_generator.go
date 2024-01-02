package resources

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	configModels "go-deploy/models/config"
	"go-deploy/models/sys/deployment"
	smModels "go-deploy/models/sys/sm"
	userModels "go-deploy/models/sys/user"
	"go-deploy/models/sys/vm"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service/constants"
	"go-deploy/utils"
	"golang.org/x/exp/slices"
	v1 "k8s.io/api/core/v1"
	"math"
	"path"
	"time"
)

type K8sGenerator struct {
	*PublicGeneratorType
	namespace string
	client    *k8s.Client
}

func (kg *K8sGenerator) Namespace() *models.NamespacePublic {
	if kg.d.deployment != nil {
		ns := models.NamespacePublic{
			Name: kg.namespace,
		}

		if n := &kg.d.deployment.Subsystems.K8s.Namespace; subsystems.Created(n) {
			ns.ID = n.ID
			ns.CreatedAt = n.CreatedAt
		}

		return &ns
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

		ns := models.NamespacePublic{
			Name: kg.namespace,
		}

		if n := &kg.v.vm.Subsystems.K8s.Namespace; subsystems.Created(n) {
			ns.ID = n.ID
			ns.CreatedAt = n.CreatedAt
		}

		return &ns
	}

	if kg.s.sm != nil {
		ns := models.NamespacePublic{
			Name: kg.namespace,
		}

		if n := &kg.s.sm.Subsystems.K8s.Namespace; subsystems.Created(n) {
			ns.ID = n.ID
			ns.CreatedAt = n.CreatedAt
		}

		return &ns
	}

	return nil
}

func (kg *K8sGenerator) Deployments() []models.DeploymentPublic {
	var res []models.DeploymentPublic

	if kg.d.deployment != nil {
		mainApp := kg.d.deployment.GetMainApp()

		var imagePullSecrets []string
		if kg.d.deployment.Type == deployment.TypeCustom {
			imagePullSecrets = []string{constants.WithImagePullSecretSuffix(kg.d.deployment.Name)}
		}

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
			CPU:    config.Config.Deployment.Resources.Limits.CPU,
			Memory: config.Config.Deployment.Resources.Limits.Memory,
		}

		defaultRequests := models.Requests{
			CPU:    config.Config.Deployment.Resources.Requests.CPU,
			Memory: config.Config.Deployment.Resources.Requests.Memory,
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

		var image string
		if mainApp.Replicas > 0 {
			image = mainApp.Image
		} else {
			image = config.Config.Registry.PlaceholderImage

			k8sEnvs = append(k8sEnvs, models.EnvVar{
				Name:  "TYPE",
				Value: "disabled",
			})
		}

		dep := models.DeploymentPublic{
			Name:             kg.d.deployment.Name,
			Namespace:        kg.namespace,
			Image:            image,
			ImagePullSecrets: imagePullSecrets,
			EnvVars:          k8sEnvs,
			Resources: models.Resources{
				Limits:   defaultLimits,
				Requests: defaultRequests,
			},
			Command:        make([]string, 0),
			Args:           make([]string, 0),
			InitCommands:   mainApp.InitCommands,
			InitContainers: make([]models.InitContainer, 0),
			Volumes:        k8sVolumes,
		}

		if d := kg.d.deployment.Subsystems.K8s.GetDeployment(kg.d.deployment.Name); subsystems.Created(d) {
			dep.ID = d.ID
			dep.CreatedAt = d.CreatedAt
		}

		res = append(res, dep)
		return res
	}

	if kg.v.vm != nil {
		ports := kg.v.vm.Ports

		for _, port := range ports {
			if port.HttpProxy == nil {
				continue
			}

			csPort := kg.v.vm.Subsystems.CS.GetPortForwardingRule(pfrName(port.Port, port.Protocol))
			if csPort == nil {
				continue
			}

			envVars := []models.EnvVar{
				{Name: "PORT", Value: "8080"},
				{Name: "VM_PORT", Value: fmt.Sprintf("%d", csPort.PublicPort)},
				{Name: "URL", Value: vpExternalURL(port.HttpProxy.Name, kg.v.deploymentZone)},
				{Name: "VM_URL", Value: kg.v.vmZone.ParentDomain},
			}

			res = append(res, models.DeploymentPublic{
				Name:             vpDeploymentName(kg.v.vm, port.HttpProxy.Name),
				Namespace:        kg.namespace,
				Image:            config.Config.Registry.VmHttpProxyImage,
				ImagePullSecrets: make([]string, 0),
				EnvVars:          envVars,
				Resources: models.Resources{
					Limits: models.Limits{
						CPU:    config.Config.Deployment.Resources.Limits.CPU,
						Memory: config.Config.Deployment.Resources.Limits.Memory,
					},
					Requests: models.Requests{
						CPU:    config.Config.Deployment.Resources.Requests.CPU,
						Memory: config.Config.Deployment.Resources.Requests.Memory,
					},
				},
				Command:        make([]string, 0),
				Args:           make([]string, 0),
				InitCommands:   make([]string, 0),
				InitContainers: make([]models.InitContainer, 0),
				Volumes:        make([]models.Volume, 0),
			})
		}

		for mapName, k8sDeployment := range kg.v.vm.Subsystems.K8s.GetDeploymentMap() {
			idx := 0
			matchedIdx := -1
			for _, port := range ports {
				if port.HttpProxy == nil {
					continue
				}

				if vpDeploymentName(kg.v.vm, port.HttpProxy.Name) == mapName {
					matchedIdx = idx
					break
				}

				idx++
			}

			if matchedIdx != -1 {
				res[idx].ID = k8sDeployment.ID
				res[idx].CreatedAt = k8sDeployment.CreatedAt
			}
		}

		return res
	}

	if kg.s.sm != nil {
		initVolumes, stdVolume := sVolumes(kg.s.sm.OwnerID)
		allVolumes := append(initVolumes, stdVolume...)

		k8sVolumes := make([]models.Volume, len(allVolumes))
		for i, volume := range allVolumes {
			pvcName := sPvcName(volume.Name)
			k8sVolumes[i] = models.Volume{
				Name:      sPvName(kg.s.sm.OwnerID, volume.Name),
				PvcName:   &pvcName,
				MountPath: volume.AppPath,
				Init:      volume.Init,
			}
		}

		defaultLimits := models.Limits{
			CPU:    config.Config.Deployment.Resources.Limits.CPU,
			Memory: config.Config.Deployment.Resources.Limits.Memory,
		}

		defaultRequests := models.Requests{
			CPU:    config.Config.Deployment.Resources.Requests.CPU,
			Memory: config.Config.Deployment.Resources.Requests.Memory,
		}

		args := []string{
			"--noauth",
			"--root=/deploy",
			"--database=/data/database.db",
			"--port=80",
		}

		filebrowser := models.DeploymentPublic{
			Name:             constants.SmAppName,
			Namespace:        kg.namespace,
			Image:            "filebrowser/filebrowser",
			ImagePullSecrets: make([]string, 0),
			EnvVars:          make([]models.EnvVar, 0),
			Resources: models.Resources{
				Limits:   defaultLimits,
				Requests: defaultRequests,
			},
			Command:        make([]string, 0),
			Args:           args,
			InitCommands:   make([]string, 0),
			InitContainers: make([]models.InitContainer, 0),
			Volumes:        k8sVolumes,
		}

		if fb := kg.s.sm.Subsystems.K8s.GetDeployment(constants.SmAppName); subsystems.Created(fb) {
			filebrowser.ID = fb.ID
			filebrowser.CreatedAt = fb.CreatedAt
		}

		res = append(res, filebrowser)

		// oauth2-proxy
		user, err := userModels.New().GetByID(kg.s.sm.OwnerID)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to get user by id when creating oauth proxy deployment public. details: %w", err))
			return nil
		}

		volumes := []models.Volume{
			{
				Name:      "oauth-proxy-config",
				PvcName:   nil,
				MountPath: "/mnt",
				Init:      false,
			},
			{
				Name:      "oauth-proxy-config",
				PvcName:   nil,
				MountPath: "/mnt/config",
				Init:      true,
			},
		}

		issuer := config.Config.Keycloak.Url + "/realms/" + config.Config.Keycloak.Realm
		redirectURL := fmt.Sprintf("https://%s.%s/oauth2/callback", kg.s.sm.OwnerID, kg.s.zone.Storage.ParentDomain)
		upstream := "http://storage-manager"

		args = []string{
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
			"--client-id=" + config.Config.Keycloak.StorageClient.ClientID,
			"--client-secret=" + config.Config.Keycloak.StorageClient.ClientSecret,
			"--cookie-secret=qHKgjlAFQBZOnGcdH5jIKV0Auzx5r8jzZenxhJnlZJg=",
			"--cookie-secure=true",
			"--ssl-insecure-skip-verify=true",
			"--insecure-oidc-allow-unverified-email=true",
			"--skip-provider-button=true",
			"--pass-authorization-header=true",
			"--ssl-upstream-insecure-skip-verify=true",
			"--code-challenge-method=S256",
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

		oauthProxy := models.DeploymentPublic{
			Name:             constants.SmAppNameAuth,
			Namespace:        kg.namespace,
			Image:            "quay.io/oauth2-proxy/oauth2-proxy:latest",
			ImagePullSecrets: make([]string, 0),
			EnvVars:          make([]models.EnvVar, 0),
			Resources: models.Resources{
				Limits:   defaultLimits,
				Requests: defaultRequests,
			},
			Command:        make([]string, 0),
			Args:           args,
			InitCommands:   make([]string, 0),
			InitContainers: initContainers,
			Volumes:        volumes,
			CreatedAt:      time.Time{},
		}

		if op := kg.s.sm.Subsystems.K8s.GetDeployment(constants.SmAppNameAuth); subsystems.Created(op) {
			oauthProxy.ID = op.ID
			oauthProxy.CreatedAt = op.CreatedAt
		}

		res = append(res, oauthProxy)
		return res
	}

	return nil
}

func (kg *K8sGenerator) Services() []models.ServicePublic {
	var res []models.ServicePublic
	if kg.d.deployment != nil {
		mainApp := kg.d.deployment.GetMainApp()

		se := models.ServicePublic{
			Name:       kg.d.deployment.Name,
			Namespace:  kg.namespace,
			Port:       mainApp.InternalPort,
			TargetPort: mainApp.InternalPort,
		}

		if k8sService := kg.d.deployment.Subsystems.K8s.GetService(kg.d.deployment.Name); subsystems.Created(k8sService) {
			se.ID = k8sService.ID
			se.CreatedAt = k8sService.CreatedAt
		}

		res = append(res, se)
		return res
	}

	if kg.v.vm != nil {
		ports := kg.v.vm.Ports

		for _, port := range ports {
			if port.HttpProxy == nil {
				continue
			}

			res = append(res, models.ServicePublic{
				Name:       vpServiceName(kg.v.vm, port.HttpProxy.Name),
				Namespace:  kg.namespace,
				Port:       8080,
				TargetPort: 8080,
			})
		}

		for mapName, svc := range kg.v.vm.Subsystems.K8s.GetServiceMap() {
			idx := 0
			matchedIdx := -1
			for _, port := range ports {
				if port.HttpProxy == nil {
					continue
				}

				if vpServiceName(kg.v.vm, port.HttpProxy.Name) == mapName {
					matchedIdx = idx
					break
				}

				idx++
			}

			if matchedIdx != -1 {
				res[idx].ID = svc.ID
				res[idx].CreatedAt = svc.CreatedAt
			}
		}

		return res
	}

	if kg.s.sm != nil {
		// filebrowser
		filebrowser := models.ServicePublic{
			Name:       constants.SmAppName,
			Namespace:  kg.namespace,
			Port:       80,
			TargetPort: 80,
		}

		if fb := kg.s.sm.Subsystems.K8s.GetService(constants.SmAppName); subsystems.Created(fb) {
			filebrowser.ID = fb.ID
			filebrowser.CreatedAt = fb.CreatedAt
		}

		res = append(res, filebrowser)

		// oauth2-proxy
		oauthProxy := models.ServicePublic{
			Name:       constants.SmAppNameAuth,
			Namespace:  kg.namespace,
			Port:       4180,
			TargetPort: 4180,
		}

		if op := kg.s.sm.Subsystems.K8s.GetService(constants.SmAppNameAuth); subsystems.Created(op) {
			oauthProxy.ID = op.ID
			oauthProxy.CreatedAt = op.CreatedAt
		}

		res = append(res, oauthProxy)

		return res
	}

	return nil
}

func (kg *K8sGenerator) Ingresses() []models.IngressPublic {
	var res []models.IngressPublic
	if kg.d.deployment != nil {
		mainApp := kg.d.deployment.GetMainApp()

		if mainApp.Private {
			return res
		}

		tlsSecret := constants.WildcardCertSecretName
		in := models.IngressPublic{
			Name:         kg.d.deployment.Name,
			Namespace:    kg.namespace,
			ServiceName:  kg.d.deployment.Name,
			ServicePort:  kg.d.deployment.GetMainApp().InternalPort,
			IngressClass: config.Config.Deployment.IngressClass,
			Hosts:        []string{getExternalFQDN(kg.d.deployment.Name, kg.d.zone)},
			Placeholder:  false,
			TlsSecret:    &tlsSecret,
			CustomCert:   nil,
		}

		if k8sIngress := kg.d.deployment.Subsystems.K8s.GetIngress(kg.d.deployment.Name); subsystems.Created(k8sIngress) {
			in.ID = k8sIngress.ID
			in.CreatedAt = k8sIngress.CreatedAt
		}

		res = append(res, in)

		if mainApp.CustomDomain != nil && mainApp.CustomDomain.Status == deployment.CustomDomainStatusReady {
			customIn := models.IngressPublic{
				ID:           "",
				Name:         fmt.Sprintf(constants.WithCustomDomainSuffix(kg.d.deployment.Name)),
				Namespace:    kg.namespace,
				ServiceName:  kg.d.deployment.Name,
				ServicePort:  mainApp.InternalPort,
				IngressClass: config.Config.Deployment.IngressClass,
				Hosts:        []string{mainApp.CustomDomain.Domain},
				Placeholder:  false,
				CreatedAt:    time.Time{},
				CustomCert: &models.CustomCert{
					ClusterIssuer: "letsencrypt-prod-deploy-http",
					CommonName:    mainApp.CustomDomain.Domain,
				},
				TlsSecret: nil,
			}

			if customK8sIngress := kg.d.deployment.Subsystems.K8s.GetIngress(constants.WithCustomDomainSuffix(kg.d.deployment.Name)); subsystems.Created(customK8sIngress) {
				customIn.ID = customK8sIngress.ID
				customIn.CreatedAt = customK8sIngress.CreatedAt
			}

			res = append(res, customIn)
		}

		return res
	}

	if kg.v.vm != nil {
		ports := kg.v.vm.Ports

		for _, port := range ports {
			if port.HttpProxy == nil {
				continue
			}

			tlsSecret := constants.WildcardCertSecretName
			res = append(res, models.IngressPublic{
				Name:         vpIngressName(kg.v.vm, port.HttpProxy.Name),
				Namespace:    kg.namespace,
				ServiceName:  vpServiceName(kg.v.vm, port.HttpProxy.Name),
				ServicePort:  8080,
				IngressClass: config.Config.Deployment.IngressClass,
				Hosts:        []string{vpExternalURL(port.HttpProxy.Name, kg.v.deploymentZone)},
				TlsSecret:    &tlsSecret,
				CustomCert:   nil,
				Placeholder:  false,
			})

			if port.HttpProxy.CustomDomain != nil {
				res = append(res, models.IngressPublic{
					Name:         vpCustomDomainIngressName(kg.v.vm, port.HttpProxy.Name),
					Namespace:    kg.namespace,
					ServiceName:  vpServiceName(kg.v.vm, port.HttpProxy.Name),
					ServicePort:  8080,
					IngressClass: config.Config.Deployment.IngressClass,
					Hosts:        []string{*port.HttpProxy.CustomDomain},
					Placeholder:  false,
					CustomCert: &models.CustomCert{
						ClusterIssuer: "letsencrypt-prod-deploy-http",
						CommonName:    *port.HttpProxy.CustomDomain,
					},
					TlsSecret: nil,
				})
			}
		}

		for mapName, ingress := range kg.v.vm.Subsystems.K8s.GetIngressMap() {
			idx := 0
			matchedIdx := -1
			for _, port := range ports {
				if port.HttpProxy == nil {
					continue
				}

				if vpIngressName(kg.v.vm, port.HttpProxy.Name) == mapName ||
					(vpCustomDomainIngressName(kg.v.vm, port.HttpProxy.Name) == mapName && port.HttpProxy.CustomDomain != nil) {
					matchedIdx = idx
					break
				}
			}

			if matchedIdx != -1 {
				res[idx].ID = ingress.ID
				res[idx].CreatedAt = ingress.CreatedAt
			}
		}

		return res
	}

	if kg.s.sm != nil {
		tlsSecret := constants.WildcardCertSecretName

		ingress := models.IngressPublic{
			Name:         constants.SmAppName,
			Namespace:    kg.namespace,
			ServiceName:  constants.SmAppNameAuth,
			ServicePort:  4180,
			IngressClass: config.Config.Deployment.IngressClass,
			Hosts:        []string{getStorageExternalFQDN(kg.s.sm.OwnerID, kg.s.zone)},
			TlsSecret:    &tlsSecret,
		}

		if i := kg.s.sm.Subsystems.K8s.GetIngress(constants.SmAppName); subsystems.Created(i) {
			ingress.ID = i.ID
			ingress.CreatedAt = i.CreatedAt
		}

		res = append(res, ingress)
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

		for _, v := range volumes {
			res = append(res, models.PvPublic{
				Name:      dPvName(kg.d.deployment, v.Name),
				Capacity:  config.Config.Deployment.Resources.Limits.Storage,
				NfsServer: kg.s.zone.Storage.NfsServer,
				NfsPath:   path.Join(kg.s.zone.Storage.NfsParentPath, v.ServerPath),
			})
		}

		for mapName, pv := range kg.d.deployment.Subsystems.K8s.GetPvMap() {
			idx := slices.IndexFunc(res, func(pv models.PvPublic) bool {
				return pv.Name == mapName
			})
			if idx != -1 {
				res[idx].ID = pv.ID
				res[idx].CreatedAt = pv.CreatedAt
			}
		}

		return res
	}

	if kg.s.sm != nil {
		initVolumes, volumes := sVolumes(kg.s.sm.OwnerID)
		allVolumes := append(initVolumes, volumes...)

		for _, v := range allVolumes {
			res = append(res, models.PvPublic{
				Name:      sPvName(kg.s.sm.OwnerID, v.Name),
				Capacity:  config.Config.Deployment.Resources.Limits.Storage,
				NfsServer: kg.s.zone.Storage.NfsServer,
				NfsPath:   path.Join(kg.s.zone.Storage.NfsParentPath, v.ServerPath),
			})
		}

		for mapName, pv := range kg.s.sm.Subsystems.K8s.GetPvMap() {
			idx := slices.IndexFunc(res, func(pv models.PvPublic) bool {
				return pv.Name == mapName
			})
			if idx != -1 {
				res[idx].ID = pv.ID
				res[idx].CreatedAt = pv.CreatedAt
			}
		}
	}

	return res
}

func (kg *K8sGenerator) PVCs() []models.PvcPublic {
	var res []models.PvcPublic

	if kg.d.deployment != nil {
		volumes := kg.d.deployment.GetMainApp().Volumes

		for _, volume := range volumes {
			res = append(res, models.PvcPublic{
				Name:      dPvcName(kg.d.deployment, volume.Name),
				Namespace: kg.namespace,
				Capacity:  config.Config.Deployment.Resources.Limits.Storage,
				PvName:    dPvName(kg.d.deployment, volume.Name),
			})
		}

		for mapName, pvc := range kg.d.deployment.Subsystems.K8s.GetPvcMap() {
			idx := slices.IndexFunc(res, func(pvc models.PvcPublic) bool {
				return pvc.Name == mapName
			})
			if idx != -1 {
				res[idx].ID = pvc.ID
				res[idx].CreatedAt = pvc.CreatedAt
			}
		}

		return res
	}

	if kg.s.sm != nil {
		initVolumes, volumes := sVolumes(kg.s.sm.OwnerID)
		allVolumes := append(initVolumes, volumes...)

		for _, volume := range allVolumes {
			res = append(res, models.PvcPublic{
				Name:      sPvcName(volume.Name),
				Namespace: kg.namespace,
				Capacity:  config.Config.Deployment.Resources.Limits.Storage,
				PvName:    sPvName(kg.s.sm.OwnerID, volume.Name),
			})
		}

		for mapName, pvc := range kg.s.sm.Subsystems.K8s.GetPvcMap() {
			idx := slices.IndexFunc(res, func(pvc models.PvcPublic) bool {
				return pvc.Name == mapName
			})
			if idx != -1 {
				res[idx].ID = pvc.ID
				res[idx].CreatedAt = pvc.CreatedAt
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
			var imagePullSecret *models.SecretPublic

			if kg.d.deployment.Subsystems.Harbor.Robot.Created() && kg.d.deployment.Type == deployment.TypeCustom {
				registry := config.Config.Registry.URL
				username := kg.d.deployment.Subsystems.Harbor.Robot.HarborName
				password := kg.d.deployment.Subsystems.Harbor.Robot.Secret

				imagePullSecret = &models.SecretPublic{
					Name:      constants.WithImagePullSecretSuffix(kg.d.deployment.Name),
					Namespace: kg.namespace,
					Type:      string(v1.SecretTypeDockerConfigJson),
					Data: map[string][]byte{
						v1.DockerConfigJsonKey: encodeDockerConfig(registry, username, password),
					},
				}
			}

			// if already exists, set the fields that are created in the subsystem
			if secret := kg.d.deployment.Subsystems.K8s.GetSecret(constants.WithImagePullSecretSuffix(kg.d.deployment.Name)); subsystems.Created(secret) {
				if imagePullSecret == nil {
					imagePullSecret = secret
				} else {
					imagePullSecret.ID = secret.ID
					imagePullSecret.CreatedAt = secret.CreatedAt
				}
			}

			if imagePullSecret != nil {
				res = append(res, *imagePullSecret)
			}
		}

		// wildcard certificate
		/// swap namespaces temporarily
		var wildcardCertSecret *models.SecretPublic

		kg.client.Namespace = config.Config.Deployment.WildcardCertSecretNamespace
		defer func() { kg.client.Namespace = kg.namespace }()

		copyFrom, err := kg.client.ReadSecret(config.Config.Deployment.WildcardCertSecretId)
		if err != nil || copyFrom == nil {
			utils.PrettyPrintError(fmt.Errorf("failed to read secret %s/%s. details: %w", config.Config.Deployment.WildcardCertSecretNamespace, config.Config.Deployment.WildcardCertSecretId, err))
		} else {
			wildcardCertSecret = &models.SecretPublic{
				Name:      constants.WildcardCertSecretName,
				Namespace: kg.namespace,
				Type:      string(v1.SecretTypeOpaque),
				Data:      copyFrom.Data,
			}
		}

		if secret := kg.d.deployment.Subsystems.K8s.GetSecret(constants.WildcardCertSecretName); subsystems.Created(secret) {
			if wildcardCertSecret == nil {
				wildcardCertSecret = secret
			} else {
				wildcardCertSecret.ID = secret.ID
				wildcardCertSecret.CreatedAt = secret.CreatedAt
			}
		}

		if wildcardCertSecret != nil {
			res = append(res, *wildcardCertSecret)
		}

		return res
	}

	if kg.v.vm != nil {
		createWildcardSecret := false
		for _, port := range kg.v.vm.Ports {
			if port.HttpProxy != nil {
				createWildcardSecret = true
				break
			}
		}

		if !createWildcardSecret {
			return nil
		}

		// wildcard certificate
		/// swap namespaces temporarily
		var wildcardCertSecret *models.SecretPublic

		kg.client.Namespace = config.Config.Deployment.WildcardCertSecretNamespace
		defer func() { kg.client.Namespace = kg.namespace }()

		copyFrom, err := kg.client.ReadSecret(config.Config.Deployment.WildcardCertSecretId)
		if err != nil || copyFrom == nil {
			utils.PrettyPrintError(fmt.Errorf("failed to read secret %s/%s. details: %w", config.Config.Deployment.WildcardCertSecretNamespace, config.Config.Deployment.WildcardCertSecretId, err))
		} else {
			wildcardCertSecret = &models.SecretPublic{
				Name:      constants.WildcardCertSecretName,
				Namespace: kg.namespace,
				Type:      string(v1.SecretTypeOpaque),
				Data:      copyFrom.Data,
			}
		}

		if secret := kg.v.vm.Subsystems.K8s.GetSecret(constants.WildcardCertSecretName); subsystems.Created(secret) {
			if wildcardCertSecret == nil {
				wildcardCertSecret = secret
			} else {
				wildcardCertSecret.ID = secret.ID
				wildcardCertSecret.CreatedAt = secret.CreatedAt
			}
		}

		if wildcardCertSecret != nil {
			res = append(res, *wildcardCertSecret)
		}

		return res
	}

	if kg.s.sm != nil {
		// wildcard certificate
		/// swap namespaces temporarily
		var wildcardCertSecret *models.SecretPublic

		kg.client.Namespace = config.Config.Deployment.WildcardCertSecretNamespace
		defer func() { kg.client.Namespace = kg.namespace }()

		copyFrom, err := kg.client.ReadSecret(config.Config.Deployment.WildcardCertSecretId)
		if err != nil || copyFrom == nil {
			utils.PrettyPrintError(fmt.Errorf("failed to read secret %s/%s. details: %w", config.Config.Deployment.WildcardCertSecretNamespace, config.Config.Deployment.WildcardCertSecretId, err))
		} else {
			wildcardCertSecret = &models.SecretPublic{
				Name:      constants.WildcardCertSecretName,
				Namespace: kg.namespace,
				Type:      string(v1.SecretTypeOpaque),
				Data:      copyFrom.Data,
			}
		}

		if secret := kg.s.sm.Subsystems.K8s.GetSecret(constants.WildcardCertSecretName); subsystems.Created(secret) {
			if wildcardCertSecret == nil {
				wildcardCertSecret = secret
			} else {
				wildcardCertSecret.ID = secret.ID
				wildcardCertSecret.CreatedAt = secret.CreatedAt
			}
		}

		if wildcardCertSecret != nil {
			res = append(res, *wildcardCertSecret)
		}

		return res
	}

	return nil
}

func (kg *K8sGenerator) Jobs() []models.JobPublic {
	var res []models.JobPublic

	if kg.s.sm != nil {
		initVolumes, _ := sVolumes(kg.s.sm.OwnerID)
		k8sVolumes := make([]models.Volume, len(initVolumes))
		for i, volume := range initVolumes {
			pvcName := sPvcName(volume.Name)
			k8sVolumes[i] = models.Volume{
				Name:      sPvName(kg.s.sm.OwnerID, volume.Name),
				PvcName:   &pvcName,
				MountPath: volume.AppPath,
				Init:      volume.Init,
			}
		}

		for _, job := range sJobs(kg.s.sm.OwnerID) {
			res = append(res, models.JobPublic{
				Name:      job.Name,
				Namespace: kg.namespace,
				Image:     job.Image,
				Command:   job.Command,
				Args:      job.Args,
				Volumes:   k8sVolumes,
			})
		}

		for _, job := range kg.s.sm.Subsystems.K8s.GetJobMap() {
			idx := slices.IndexFunc(res, func(j models.JobPublic) bool { return j.Name == job.Name })
			if idx != -1 {
				res[idx].ID = job.ID
				res[idx].CreatedAt = job.CreatedAt
			}
		}

		return res
	}

	return nil
}

func (kg *K8sGenerator) HPAs() []models.HpaPublic {
	var res []models.HpaPublic

	if kg.d.deployment != nil {
		mainApp := kg.d.deployment.GetMainApp()

		minReplicas := 1
		maxReplicas := int(math.Max(float64(mainApp.Replicas), float64(minReplicas)))

		hpa := models.HpaPublic{
			Name:        kg.d.deployment.Name,
			Namespace:   kg.namespace,
			MinReplicas: minReplicas,
			MaxReplicas: maxReplicas,
			Target: models.Target{
				Kind:       "Deployment",
				Name:       kg.d.deployment.Name,
				ApiVersion: "apps/v1",
			},
			CpuAverageUtilization:    config.Config.Deployment.Resources.AutoScale.CpuThreshold,
			MemoryAverageUtilization: config.Config.Deployment.Resources.AutoScale.MemoryThreshold,
		}

		if h := kg.d.deployment.Subsystems.K8s.GetHPA(kg.d.deployment.Name); subsystems.Created(h) {
			hpa.ID = h.ID
			hpa.CreatedAt = h.CreatedAt
		}

		res = append(res, hpa)
		return res
	}

	return nil
}

func getExternalFQDN(name string, zone *configModels.DeploymentZone) string {
	return fmt.Sprintf("%s.%s", name, zone.ParentDomain)
}

func getStorageExternalFQDN(name string, zone *configModels.DeploymentZone) string {
	return fmt.Sprintf("%s.%s", name, zone.Storage.ParentDomain)
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

// deployment pv name
func dPvName(deployment *deployment.Deployment, volumeName string) string {
	return fmt.Sprintf("%s-%s", deployment.Name, volumeName)
}

// deployment pvc name
func dPvcName(deployment *deployment.Deployment, volumeName string) string {
	return fmt.Sprintf("%s-%s", deployment.Name, volumeName)
}

// vm proxy deployment name
func vpDeploymentName(vm *vm.VM, portName string) string {
	return fmt.Sprintf("%s-%s", vm.Name, portName)
}

// vm proxy service name
func vpServiceName(vm *vm.VM, portName string) string {
	return fmt.Sprintf("%s-%s", vm.Name, portName)
}

// vm proxy ingress name
func vpIngressName(vm *vm.VM, portName string) string {
	return fmt.Sprintf("%s-%s", vm.Name, portName)
}

// vm proxy custom domain ingress name
func vpCustomDomainIngressName(vm *vm.VM, portName string) string {
	return fmt.Sprintf("%s-%s-custom-domain", vm.Name, portName)
}

// vm proxy external url
func vpExternalURL(portName string, zone *configModels.DeploymentZone) string {
	return fmt.Sprintf("%s.%s", portName, zone.ParentDomainVM)
}

// storage manager pvc name
func sPvcName(volumeName string) string {
	return fmt.Sprintf("%s-%s", constants.SmAppName, volumeName)
}

// storage manager pv name
func sPvName(ownerID, volumeName string) string {
	return fmt.Sprintf("%s-%s", volumeName, ownerID)
}

// storage manager volumes
func sVolumes(ownerID string) ([]smModels.Volume, []smModels.Volume) {
	initVolumes := []smModels.Volume{
		{
			Name:       "init",
			Init:       false,
			AppPath:    "/exports",
			ServerPath: "",
		},
	}

	volumes := []smModels.Volume{
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

// storage manager jobs
func sJobs(userID string) []smModels.Job {
	return []smModels.Job{
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

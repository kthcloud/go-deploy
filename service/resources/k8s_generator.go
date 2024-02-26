package resources

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	configModels "go-deploy/models/config"
	"go-deploy/models/sys/deployment"
	smModels "go-deploy/models/sys/sm"
	userModels "go-deploy/models/sys/user"
	vmModels "go-deploy/models/sys/vm"
	"go-deploy/models/versions"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/pkg/subsystems/k8s/keys"
	"go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service/constants"
	"go-deploy/utils"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	"math"
	"path"
	regexp "regexp"
	"strings"
)

// K8sGenerator is a generator for K8s resources
// It is used to generate the `publics`, such as models.DeploymentPublic and models.IngressPublic
type K8sGenerator struct {
	*PublicGeneratorType
	namespace           string
	client              *k8s.Client
	extraAuthorizedKeys []string
}

type CloudInit struct {
	FQDN            string          `yaml:"fqdn"`
	Users           []CloudInitUser `yaml:"users"`
	SshPasswordAuth bool            `yaml:"ssh_pwauth"`
	RunCMD          []string        `yaml:"runcmd"`
}

type CloudInitUser struct {
	Name              string   `yaml:"name"`
	Sudo              []string `yaml:"sudo"`
	Passwd            string   `yaml:"passwd"`
	LockPasswd        bool     `yaml:"lock_passwd"`
	Shell             string   `yaml:"shell"`
	SshAuthorizedKeys []string `yaml:"ssh_authorized_keys"`
}

func (kg *K8sGenerator) WithAuthorizedKeys(keys ...string) *K8sGenerator {
	kg.extraAuthorizedKeys = keys
	return kg
}

// Namespace returns a models.NamespacePublic that should be created
func (kg *K8sGenerator) Namespace() *models.NamespacePublic {
	if kg.d.deployment != nil {
		ns := models.NamespacePublic{
			Name: kg.namespace,
		}

		if n := &kg.d.deployment.Subsystems.K8s.Namespace; subsystems.Created(n) {
			ns.CreatedAt = n.CreatedAt
		}

		return &ns
	}

	if kg.v.vm != nil && kg.v.vm.Version == versions.V1 {
		createNamespace := false
		for _, port := range kg.v.vm.PortMap {
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
			ns.CreatedAt = n.CreatedAt
		}

		return &ns
	}

	if kg.v.vm != nil && kg.v.vm.Version == versions.V2 {
		ns := models.NamespacePublic{
			Name: kg.namespace,
		}

		if n := &kg.v.vm.Subsystems.K8s.Namespace; subsystems.Created(n) {
			ns.CreatedAt = n.CreatedAt
		}

		return &ns
	}

	if kg.s.sm != nil {
		ns := models.NamespacePublic{
			Name: kg.namespace,
		}

		if n := &kg.s.sm.Subsystems.K8s.Namespace; subsystems.Created(n) {
			ns.CreatedAt = n.CreatedAt
		}

		return &ns
	}

	return nil
}

// Deployments returns a list of models.DeploymentPublic that should be created
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
			pvcName := fmt.Sprintf("%s-%s", kg.d.deployment.Name, makeValidK8sName(volume.Name))
			k8sVolumes[i] = models.Volume{
				Name:      makeValidK8sName(volume.Name),
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
			Args:           mainApp.Args,
			InitCommands:   mainApp.InitCommands,
			InitContainers: make([]models.InitContainer, 0),
			Volumes:        k8sVolumes,
		}

		if d := kg.d.deployment.Subsystems.K8s.GetDeployment(kg.d.deployment.Name); subsystems.Created(d) {
			dep.CreatedAt = d.CreatedAt
		}

		res = append(res, dep)
		return res
	}

	if kg.v.vm != nil && kg.v.vm.Version == versions.V1 {
		portMap := kg.v.vm.PortMap

		for _, port := range portMap {
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
				{Name: "URL", Value: vmpExternalURL(port.HttpProxy.Name, kg.v.deploymentZone)},
				{Name: "VM_URL", Value: kg.v.vmZone.ParentDomain},
			}

			res = append(res, models.DeploymentPublic{
				Name:             vmpDeploymentName(kg.v.vm, port.HttpProxy.Name),
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
			for _, port := range portMap {
				if port.HttpProxy == nil {
					continue
				}

				if vmpDeploymentName(kg.v.vm, port.HttpProxy.Name) == mapName {
					matchedIdx = idx
					break
				}

				idx++
			}

			if matchedIdx != -1 {
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
			pvcName := sPvcName(kg.s.sm.OwnerID, volume.Name)
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

		// Filebrowser
		args := []string{
			"--noauth",
			"--root=/deploy",
			"--database=/data/database.db",
			"--port=80",
		}

		filebrowser := models.DeploymentPublic{
			Name:             smName(kg.s.sm.OwnerID),
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

		if fb := kg.s.sm.Subsystems.K8s.GetDeployment(smName(kg.s.sm.OwnerID)); subsystems.Created(fb) {
			filebrowser.CreatedAt = fb.CreatedAt
		}

		res = append(res, filebrowser)

		// Oauth2-proxy
		user, err := userModels.New().GetByID(kg.s.sm.OwnerID)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to get user by id when creating oauth proxy deployment public. details: %w", err))
			return nil
		}

		volumes := []models.Volume{
			{
				Name:      "oauth-proxy-config",
				MountPath: "/mnt",
				Init:      false,
			},
			{
				Name:      "oauth-proxy-config",
				MountPath: "/mnt/config",
				Init:      true,
			},
		}

		issuer := config.Config.Keycloak.Url + "/realms/" + config.Config.Keycloak.Realm
		redirectURL := fmt.Sprintf("https://%s.%s/oauth2/callback", kg.s.sm.OwnerID, kg.s.zone.Storage.ParentDomain)
		upstream := "http://" + smName(kg.s.sm.OwnerID) + ":80"

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

		oauthProxy := models.DeploymentPublic{
			Name:             smAuthName(kg.s.sm.OwnerID),
			Namespace:        kg.namespace,
			Image:            "quay.io/oauth2-proxy/oauth2-proxy:latest",
			ImagePullSecrets: make([]string, 0),
			EnvVars:          make([]models.EnvVar, 0),
			Resources: models.Resources{
				Limits:   defaultLimits,
				Requests: defaultRequests,
			},
			Command:      make([]string, 0),
			Args:         args,
			InitCommands: make([]string, 0),
			InitContainers: []models.InitContainer{{
				Name:    "oauth-proxy-config-init",
				Image:   "busybox",
				Command: []string{"sh", "-c", fmt.Sprintf("mkdir -p /mnt/config && echo %s > /mnt/config/authenticated-emails-list", user.Email)},
				Args:    nil,
			}},
			Volumes: volumes,
		}

		if op := kg.s.sm.Subsystems.K8s.GetDeployment(smAuthName(kg.s.sm.OwnerID)); subsystems.Created(op) {
			oauthProxy.CreatedAt = op.CreatedAt
		}

		res = append(res, oauthProxy)
		return res
	}

	return nil
}

func (kg *K8sGenerator) VMs() []models.VmPublic {
	var res []models.VmPublic

	if kg.v.vm != nil && kg.v.vm.Version == versions.V2 {
		sshPublicKeys := make([]string, len(kg.extraAuthorizedKeys)+1)
		sshPublicKeys[0] = kg.v.vm.SshPublicKey
		copy(sshPublicKeys[1:], kg.extraAuthorizedKeys)

		cloudInit := CloudInit{
			FQDN: kg.v.vm.Name,
			Users: []CloudInitUser{
				{
					Name:              "root",
					Sudo:              []string{"ALL=(ALL) NOPASSWD:ALL"},
					Passwd:            utils.HashPassword("root", utils.GenerateSalt()),
					LockPasswd:        false,
					Shell:             "/bin/bash",
					SshAuthorizedKeys: sshPublicKeys,
				},
			},
			SshPasswordAuth: false,
			RunCMD:          []string{"git clone https://github.com/kthcloud/boostrap-vm.git init && cd init && chmod +x run.sh && ./run.sh"},
		}

		vmPublic := models.VmPublic{
			Name:      vmName(kg.v.vm),
			Namespace: kg.namespace,

			CpuCores: kg.v.vm.Specs.CpuCores,
			RAM:      kg.v.vm.Specs.RAM,
			DiskSize: kg.v.vm.Specs.DiskSize,

			CloudInit: createCloudInitString(&cloudInit),
			// Temporary image URL
			Image: "docker://registry.cloud.cbh.kth.se/images/ubuntu:24.04",
		}

		if vm := kg.v.vm.Subsystems.K8s.GetVM(vmName(kg.v.vm)); subsystems.Created(vm) {
			vmPublic.ID = vm.ID
			vmPublic.Running = vm.Running
			vmPublic.CreatedAt = vm.CreatedAt
		}

		res = append(res, vmPublic)
		return res
	}

	return nil
}

// Services returns a list of models.ServicePublic that should be created
func (kg *K8sGenerator) Services() []models.ServicePublic {
	var res []models.ServicePublic
	if kg.d.deployment != nil {
		mainApp := kg.d.deployment.GetMainApp()

		se := models.ServicePublic{
			Name:      kg.d.deployment.Name,
			Namespace: kg.namespace,
			Ports:     []models.Port{{Name: "http", Protocol: "tcp", Port: mainApp.InternalPort, TargetPort: mainApp.InternalPort}},
			Selector: map[string]string{
				keys.LabelDeployName: kg.d.deployment.Name,
			},
		}

		if k8sService := kg.d.deployment.Subsystems.K8s.GetService(kg.d.deployment.Name); subsystems.Created(k8sService) {
			se.CreatedAt = k8sService.CreatedAt
		}

		res = append(res, se)
		return res
	}

	if kg.v.vm != nil && kg.v.vm.Version == versions.V1 {
		portMap := kg.v.vm.PortMap

		for _, port := range portMap {
			if port.HttpProxy == nil {
				continue
			}

			res = append(res, models.ServicePublic{
				Name:      vmpServiceName(kg.v.vm, port.HttpProxy.Name),
				Namespace: kg.namespace,
				Ports:     []models.Port{{Name: pfrName(port.Port, port.Protocol), Protocol: port.Protocol, Port: 8080, TargetPort: 8080}},
				Selector: map[string]string{
					keys.LabelDeployName: vmpDeploymentName(kg.v.vm, port.HttpProxy.Name),
				},
			})
		}

		for mapName, svc := range kg.v.vm.Subsystems.K8s.GetServiceMap() {
			idx := 0
			matchedIdx := -1
			for _, port := range portMap {
				if port.HttpProxy == nil {
					continue
				}

				if vmpServiceName(kg.v.vm, port.HttpProxy.Name) == mapName {
					matchedIdx = idx
					break
				}

				idx++
			}

			if matchedIdx != -1 {
				res[idx].CreatedAt = svc.CreatedAt
			}
		}

		return res
	}

	if kg.v.vm != nil && kg.v.vm.Version == versions.V2 {
		portMap := kg.v.vm.PortMap

		for _, port := range portMap {
			res = append(res, models.ServicePublic{
				Name:      vmServiceName(kg.v.vm),
				Namespace: kg.namespace,
				Ports: []models.Port{{
					Name:       pfrName(port.Port, port.Protocol),
					Protocol:   port.Protocol,
					Port:       0, // This is set externally
					TargetPort: port.Port,
				}},
				Selector: map[string]string{
					keys.LabelDeployName: vmName(kg.v.vm),
				},
				LoadBalancerIP: &kg.v.deploymentZone.LoadBalancerIP,
			})
		}

		for mapName, s := range kg.v.vm.Subsystems.K8s.GetServiceMap() {
			idx := slices.IndexFunc(res, func(service models.ServicePublic) bool {
				return service.Name == mapName
			})
			if idx != -1 {
				// Set external ports
				for _, port := range portMap {
					for _, p := range res[idx].Ports {
						if p.Name == pfrName(port.Port, port.Protocol) {
							res[idx].Ports[idx].Port = s.Ports[0].Port
							break
						}
					}
				}

				res[idx].CreatedAt = s.CreatedAt
			}
		}

		// TODO: Add services for proxy ports

		return res
	}

	if kg.s.sm != nil {
		// Filebrowser
		filebrowser := models.ServicePublic{
			Name:      smName(kg.s.sm.OwnerID),
			Namespace: kg.namespace,
			Ports:     []models.Port{{Name: "http", Protocol: "tcp", Port: 80, TargetPort: 80}},
			Selector: map[string]string{
				keys.LabelDeployName: smName(kg.s.sm.OwnerID),
			},
		}

		if fb := kg.s.sm.Subsystems.K8s.GetService(smName(kg.s.sm.OwnerID)); subsystems.Created(fb) {
			filebrowser.CreatedAt = fb.CreatedAt
		}

		res = append(res, filebrowser)

		// Oauth2-proxy
		oauthProxy := models.ServicePublic{
			Name:      smAuthName(kg.s.sm.OwnerID),
			Namespace: kg.namespace,
			Ports:     []models.Port{{Name: "http", Protocol: "tcp", Port: 4180, TargetPort: 4180}},
			Selector: map[string]string{
				keys.LabelDeployName: smAuthName(kg.s.sm.OwnerID),
			},
		}

		if op := kg.s.sm.Subsystems.K8s.GetService(smAuthName(kg.s.sm.OwnerID)); subsystems.Created(op) {
			oauthProxy.CreatedAt = op.CreatedAt
		}

		res = append(res, oauthProxy)

		return res
	}

	return nil
}

// Ingresses returns a list of models.IngressPublic that should be created
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
			in.CreatedAt = k8sIngress.CreatedAt
		}

		res = append(res, in)

		if mainApp.CustomDomain != nil && mainApp.CustomDomain.Status == deployment.CustomDomainStatusActive {
			customIn := models.IngressPublic{
				Name:         fmt.Sprintf(constants.WithCustomDomainSuffix(kg.d.deployment.Name)),
				Namespace:    kg.namespace,
				ServiceName:  kg.d.deployment.Name,
				ServicePort:  mainApp.InternalPort,
				IngressClass: config.Config.Deployment.IngressClass,
				Hosts:        []string{mainApp.CustomDomain.Domain},
				CustomCert: &models.CustomCert{
					ClusterIssuer: "letsencrypt-prod-deploy-http",
					CommonName:    mainApp.CustomDomain.Domain,
				},
				TlsSecret: nil,
			}

			if customK8sIngress := kg.d.deployment.Subsystems.K8s.GetIngress(constants.WithCustomDomainSuffix(kg.d.deployment.Name)); subsystems.Created(customK8sIngress) {
				customIn.CreatedAt = customK8sIngress.CreatedAt
			}

			res = append(res, customIn)
		}

		return res
	}

	if kg.v.vm != nil && kg.v.vm.Version == versions.V1 {
		portMap := kg.v.vm.PortMap

		for _, port := range portMap {
			if port.HttpProxy == nil {
				continue
			}

			tlsSecret := constants.WildcardCertSecretName
			res = append(res, models.IngressPublic{
				Name:         vmpIngressName(kg.v.vm, port.HttpProxy.Name),
				Namespace:    kg.namespace,
				ServiceName:  vmpServiceName(kg.v.vm, port.HttpProxy.Name),
				ServicePort:  8080,
				IngressClass: config.Config.Deployment.IngressClass,
				Hosts:        []string{vmpExternalURL(port.HttpProxy.Name, kg.v.deploymentZone)},
				TlsSecret:    &tlsSecret,
				CustomCert:   nil,
				Placeholder:  false,
			})
			if port.HttpProxy.CustomDomain != nil && port.HttpProxy.CustomDomain.Status == deployment.CustomDomainStatusActive {
				res = append(res, models.IngressPublic{
					Name:         vmpCustomDomainIngressName(kg.v.vm, port.HttpProxy.Name),
					Namespace:    kg.namespace,
					ServiceName:  vmpServiceName(kg.v.vm, port.HttpProxy.Name),
					ServicePort:  8080,
					IngressClass: config.Config.Deployment.IngressClass,
					Hosts:        []string{port.HttpProxy.CustomDomain.Domain},
					Placeholder:  false,
					CustomCert: &models.CustomCert{
						ClusterIssuer: "letsencrypt-prod-deploy-http",
						CommonName:    port.HttpProxy.CustomDomain.Domain,
					},
					TlsSecret: nil,
				})
			}
		}

		for mapName, ingress := range kg.v.vm.Subsystems.K8s.GetIngressMap() {
			idx := 0
			matchedIdx := -1
			for _, port := range portMap {
				if port.HttpProxy == nil {
					continue
				}

				if vmpIngressName(kg.v.vm, port.HttpProxy.Name) == mapName ||
					(vmpCustomDomainIngressName(kg.v.vm, port.HttpProxy.Name) == mapName && port.HttpProxy.CustomDomain != nil) {
					matchedIdx = idx
					break
				}
			}

			if matchedIdx != -1 {
				res[idx].CreatedAt = ingress.CreatedAt
			}
		}

		return res
	}

	// TODO: add ingresses for VM v2

	if kg.s.sm != nil {
		tlsSecret := constants.WildcardCertSecretName

		ingress := models.IngressPublic{
			Name:         smName(kg.s.sm.OwnerID),
			Namespace:    kg.namespace,
			ServiceName:  smAuthName(kg.s.sm.OwnerID),
			ServicePort:  4180,
			IngressClass: config.Config.Deployment.IngressClass,
			Hosts:        []string{getStorageExternalFQDN(kg.s.sm.OwnerID, kg.s.zone)},
			TlsSecret:    &tlsSecret,
		}

		if i := kg.s.sm.Subsystems.K8s.GetIngress(smName(kg.s.sm.OwnerID)); subsystems.Created(i) {
			ingress.CreatedAt = i.CreatedAt
		}

		res = append(res, ingress)
		return res
	}

	return nil
}

// PrivateIngress returns a models.IngressPublic that should be created
func (kg *K8sGenerator) PrivateIngress() *models.IngressPublic {
	return &models.IngressPublic{
		Placeholder: true,
	}
}

// PVs returns a list of models.PvPublic that should be created
func (kg *K8sGenerator) PVs() []models.PvPublic {
	var res []models.PvPublic

	if kg.d.deployment != nil {
		volumes := kg.d.deployment.GetMainApp().Volumes

		for _, v := range volumes {
			res = append(res, models.PvPublic{
				Name:      dPvName(kg.d.deployment, v.Name),
				Capacity:  config.Config.Deployment.Resources.Limits.Storage,
				NfsServer: kg.s.zone.Storage.NfsServer,
				// Create path /<zone parent path>/<deployment owner id>/user/<user specified server path>
				NfsPath: path.Join(kg.s.zone.Storage.NfsParentPath, kg.d.deployment.OwnerID, "user", v.ServerPath),
			})
		}

		for mapName, pv := range kg.d.deployment.Subsystems.K8s.GetPvMap() {
			idx := slices.IndexFunc(res, func(pv models.PvPublic) bool {
				return pv.Name == mapName
			})
			if idx != -1 {
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
				res[idx].CreatedAt = pv.CreatedAt
			}
		}
	}

	return res
}

// PVCs returns a list of models.PvcPublic that should be created
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
				res[idx].CreatedAt = pvc.CreatedAt
			}
		}

		return res
	}

	if kg.v.vm != nil && kg.v.vm.Version == versions.V2 {
		parentPVC := models.PvcPublic{
			Name:      vmParentPvName(kg.v.vm),
			Namespace: kg.namespace,
			Capacity:  config.Config.Deployment.Resources.Limits.Storage,
			PvName:    vmParentPvName(kg.v.vm),
		}

		if pvc := kg.v.vm.Subsystems.K8s.GetPV(vmParentPvName(kg.v.vm)); subsystems.Created(pvc) {
			parentPVC.CreatedAt = pvc.CreatedAt
		}

		return []models.PvcPublic{parentPVC}
	}

	if kg.s.sm != nil {
		initVolumes, volumes := sVolumes(kg.s.sm.OwnerID)
		allVolumes := append(initVolumes, volumes...)

		for _, volume := range allVolumes {
			res = append(res, models.PvcPublic{
				Name:      sPvcName(kg.s.sm.OwnerID, volume.Name),
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
				res[idx].CreatedAt = pvc.CreatedAt
			}
		}

		return res
	}

	return res
}

// Secrets returns a list of models.SecretPublic that should be created
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

		copyFrom, err := kg.client.ReadSecret(config.Config.Deployment.WildcardCertSecretName)
		if err != nil || copyFrom == nil {
			utils.PrettyPrintError(fmt.Errorf("failed to read secret %s/%s. details: %w", config.Config.Deployment.WildcardCertSecretNamespace, config.Config.Deployment.WildcardCertSecretName, err))
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
		for _, port := range kg.v.vm.PortMap {
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

		copyFrom, err := kg.client.ReadSecret(config.Config.Deployment.WildcardCertSecretName)
		if err != nil || copyFrom == nil {
			utils.PrettyPrintError(fmt.Errorf("failed to read secret %s/%s. details: %w", config.Config.Deployment.WildcardCertSecretNamespace, config.Config.Deployment.WildcardCertSecretName, err))
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

		copyFrom, err := kg.client.ReadSecret(config.Config.Deployment.WildcardCertSecretName)
		if err != nil || copyFrom == nil {
			utils.PrettyPrintError(fmt.Errorf("failed to read secret %s/%s. details: %w", config.Config.Deployment.WildcardCertSecretNamespace, config.Config.Deployment.WildcardCertSecretName, err))
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

// Jobs returns a list of models.JobPublic that should be created
func (kg *K8sGenerator) Jobs() []models.JobPublic {
	var res []models.JobPublic

	if kg.s.sm != nil {
		// These are assumed to be one-shot jobs

		initVolumes, _ := sVolumes(kg.s.sm.OwnerID)
		k8sVolumes := make([]models.Volume, len(initVolumes))
		for i, volume := range initVolumes {
			pvcName := sPvcName(kg.s.sm.OwnerID, volume.Name)
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

		return res
	}

	return nil
}

// HPAs returns a list of models.HpaPublic that should be created
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
			hpa.CreatedAt = h.CreatedAt
		}

		res = append(res, hpa)
		return res
	}

	return nil
}

// NetworkPolicies returns a list of models.NetworkPolicyPublic that should be created
func (kg *K8sGenerator) NetworkPolicies() []models.NetworkPolicyPublic {
	var res []models.NetworkPolicyPublic

	if kg.d.deployment != nil {
		for _, egressRule := range kg.d.zone.NetworkPolicies {
			egressRules := make([]models.EgressRule, 0)
			for _, egress := range egressRule.Egress {
				egressRules = append(egressRules, models.EgressRule{
					IpBlock: &models.IpBlock{
						CIDR:   egress.IP.Allow,
						Except: egress.IP.Except,
					},
				})
			}

			np := models.NetworkPolicyPublic{
				Name:        networkPolicyName(kg.d.deployment.Name, egressRule.Name),
				Namespace:   kg.namespace,
				Selector:    map[string]string{keys.LabelDeployName: kg.d.deployment.Name},
				EgressRules: egressRules,
				IngressRules: []models.IngressRule{
					{
						PodSelector:       map[string]string{"owner-id": kg.d.deployment.OwnerID},
						NamespaceSelector: map[string]string{"kubernetes.io/metadata.name": kg.namespace},
					},
					{
						NamespaceSelector: map[string]string{"kubernetes.io/metadata.name": kg.d.zone.IngressNamespace},
					},
				},
			}

			if npo := kg.d.deployment.Subsystems.K8s.GetNetworkPolicy(egressRule.Name); subsystems.Created(npo) {
				np.CreatedAt = npo.CreatedAt
			}

			res = append(res, np)
		}

		return res
	}

	if kg.v.vm != nil && kg.v.vm.Version == versions.V1 {
		if !anyHttpProxy(kg.v.vm) {
			return nil
		}

		for _, egressRule := range kg.v.deploymentZone.NetworkPolicies {
			egressRules := make([]models.EgressRule, 0)
			for _, egress := range egressRule.Egress {
				egressRules = append(egressRules, models.EgressRule{
					IpBlock: &models.IpBlock{
						CIDR:   egress.IP.Allow,
						Except: egress.IP.Except,
					},
				})
			}

			np := models.NetworkPolicyPublic{
				Name:        networkPolicyName(kg.v.vm.Name, egressRule.Name),
				Namespace:   kg.namespace,
				Selector:    map[string]string{keys.LabelDeployName: kg.v.vm.Name},
				EgressRules: egressRules,
				IngressRules: []models.IngressRule{
					{
						PodSelector:       map[string]string{"owner-id": kg.v.vm.OwnerID},
						NamespaceSelector: map[string]string{"kubernetes.io/metadata.name": kg.namespace},
					},
					{
						NamespaceSelector: map[string]string{"kubernetes.io/metadata.name": kg.v.deploymentZone.IngressNamespace},
					},
				},
			}

			if npo := kg.v.vm.Subsystems.K8s.GetNetworkPolicy(egressRule.Name); subsystems.Created(npo) {
				np.CreatedAt = npo.CreatedAt
			}

			res = append(res, np)
		}

		return res
	}

	return nil

}

// getExternalFQDN returns the external FQDN for a deployment in a given zone
func getExternalFQDN(name string, zone *configModels.DeploymentZone) string {
	return fmt.Sprintf("%s.%s", name, zone.ParentDomain)
}

// getStorageExternalFQDN returns the external FQDN for a storage manager in a given zone
func getStorageExternalFQDN(name string, zone *configModels.DeploymentZone) string {
	return fmt.Sprintf("%s.%s", name, zone.Storage.ParentDomain)
}

// encodeDockerConfig encodes docker config to json to be able to use it as a secret
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

// smName returns the name for a storage manager
func smName(userID string) string {
	return fmt.Sprintf("%s-%s", constants.SmAppName, userID)
}

// smAuthName returns the name for a storage manager auth proxy
func smAuthName(userID string) string {
	return fmt.Sprintf("%s-%s", constants.SmAppNameAuth, userID)
}

// dPvName returns the PV name for a deployment
func dPvName(deployment *deployment.Deployment, volumeName string) string {
	return fmt.Sprintf("%s-%s", deployment.Name, makeValidK8sName(volumeName))
}

// dPvcName returns the PVC name for a deployment
func dPvcName(deployment *deployment.Deployment, volumeName string) string {
	return fmt.Sprintf("%s-%s", deployment.Name, makeValidK8sName(volumeName))
}

// vmName returns the VM name for a VM
func vmName(vm *vmModels.VM) string {
	return vm.Name
}

// vmParentPvName returns the PV name for a VM
func vmParentPvName(vm *vmModels.VM) string {
	return fmt.Sprintf("%s-%s", vm.Name, constants.VmParentName)
}

// vmServiceName returns the service name for a VM
func vmServiceName(vm *vmModels.VM) string {
	return vm.Name
}

// vmpDeploymentName returns the deployment name for a VM proxy
func vmpDeploymentName(vm *vmModels.VM, portName string) string {
	return fmt.Sprintf("%s-%s", vm.Name, portName)
}

// vmpServiceName returns the service name for a VM proxy
func vmpServiceName(vm *vmModels.VM, portName string) string {
	return fmt.Sprintf("%s-%s", vm.Name, portName)
}

// vmpIngressName returns the ingress name for a VM proxy
func vmpIngressName(vm *vmModels.VM, portName string) string {
	return fmt.Sprintf("%s-%s", vm.Name, portName)
}

// vmpCustomDomainIngressName returns the ingress name for a VM proxy custom domain
func vmpCustomDomainIngressName(vm *vmModels.VM, portName string) string {
	return fmt.Sprintf("%s-%s-custom-domain", vm.Name, portName)
}

// vmpExternalURL returns the external URL for a VM proxy
func vmpExternalURL(portName string, zone *configModels.DeploymentZone) string {
	return fmt.Sprintf("%s.%s", portName, zone.ParentDomainVmHttpProxy)
}

// sPvcName returns the PVC name for a storage manager
func sPvcName(ownerID, volumeName string) string {
	return fmt.Sprintf("sm-%s-%s", volumeName, ownerID)
}

// sPvName returns the PV name for a storage manager
func sPvName(ownerID, volumeName string) string {
	return fmt.Sprintf("sm-%s-%s", volumeName, ownerID)
}

// networkPolicyName returns the network policy name for a VM or Deployment
func networkPolicyName(name, egressRuleName string) string {
	return fmt.Sprintf("%s-%s", name, egressRuleName)
}

// sVolumes returns the init and standard volumes for a storage manager
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

// sJobs returns the init jobs for a storage manager
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

// makeValidK8sName returns a valid Kubernetes name
// It returns a string that conforms to the Kubernetes naming convention (RFC 1123)
func makeValidK8sName(name string) string {
	// Define a regular expression for invalid Kubernetes name characters
	re := regexp.MustCompile(`[^a-zA-Z0-9-.]`)

	// Replace invalid characters with '-'
	validName := re.ReplaceAllString(name, "-")

	// Convert to lower case
	validName = strings.ToLower(validName)

	// Ensure it starts and ends with an alphanumeric character
	validName = strings.TrimLeft(validName, "-.")
	validName = strings.TrimRight(validName, "-.")

	// Kubernetes names are typically limited to 253 characters
	if len(validName) > 253 {
		validName = validName[:253]
	}

	return validName
}

// createCloudInitString creates a cloud-init string from a cloud-init struct
func createCloudInitString(cloudInit *CloudInit) string {
	// Convert cloud-init struct to yaml
	yamlBytes, err := yaml.Marshal(cloudInit)
	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("failed to marshal cloud-init struct to yaml. details: %w", err))
		return ""
	}

	return "#cloud-config\n" + string(yamlBytes)
}

// anyHttpProxy returns true if a VM has any HTTP proxy ports
func anyHttpProxy(vm *vmModels.VM) bool {
	for _, port := range vm.PortMap {
		if port.HttpProxy != nil {
			return true
		}
	}

	return false
}

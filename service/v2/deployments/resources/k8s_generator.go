package resources

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	configModels "go-deploy/models/config"
	"go-deploy/models/model"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/pkg/subsystems/k8s/keys"
	"go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service/constants"
	"go-deploy/service/generators"
	"go-deploy/utils"
	v1 "k8s.io/api/core/v1"
	"math"
	"path"
	"regexp"
	"slices"
	"strings"
	"time"
)

type K8sGenerator struct {
	generators.K8sGeneratorBase

	namespace string
	client    *k8s.Client

	deployment *model.Deployment
	zone       *configModels.Zone
}

func K8s(deployment *model.Deployment, zone *configModels.Zone, client *k8s.Client, namespace string) *K8sGenerator {
	return &K8sGenerator{
		namespace:  namespace,
		client:     client,
		deployment: deployment,
		zone:       zone,
	}
}

func (kg *K8sGenerator) Namespace() *models.NamespacePublic {
	ns := models.NamespacePublic{
		Name: kg.namespace,
	}

	if n := &kg.deployment.Subsystems.K8s.Namespace; subsystems.Created(n) {
		ns.CreatedAt = n.CreatedAt
	}

	return &ns
}

func (kg *K8sGenerator) Deployments() []models.DeploymentPublic {
	mainApp := kg.deployment.GetMainApp()

	var imagePullSecrets []string
	if kg.deployment.Type == model.DeploymentTypeCustom {
		imagePullSecrets = []string{constants.WithImagePullSecretSuffix(kg.deployment.Name)}
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

	limits := models.Limits{
		CPU:    fmt.Sprintf("%.1f", mainApp.CpuCores),
		Memory: fmt.Sprintf("%dMi", int(mainApp.RAM*1000)),
	}

	requests := models.Requests{
		CPU:    fmt.Sprintf("%.1f", math.Min(config.Config.Deployment.Resources.Requests.CPU, mainApp.CpuCores)),
		Memory: fmt.Sprintf("%dMi", int(math.Min(config.Config.Deployment.Resources.Requests.RAM, mainApp.RAM)*1000)),
	}

	k8sVolumes := make([]models.Volume, len(mainApp.Volumes))
	for i, volume := range mainApp.Volumes {
		pvcName := fmt.Sprintf("%s-%s", kg.deployment.Name, makeValidK8sName(volume.Name))
		k8sVolumes[i] = models.Volume{
			Name:      makeValidK8sName(volume.Name),
			PvcName:   &pvcName,
			MountPath: volume.AppPath,
			Init:      volume.Init,
		}
	}

	dep := models.DeploymentPublic{
		Name:             kg.deployment.Name,
		Namespace:        kg.namespace,
		Labels:           map[string]string{"owner-id": kg.deployment.OwnerID},
		Image:            mainApp.Image,
		ImagePullSecrets: imagePullSecrets,
		EnvVars:          k8sEnvs,
		Resources: models.Resources{
			Limits:   limits,
			Requests: requests,
		},
		Command:        make([]string, 0),
		Args:           mainApp.Args,
		InitCommands:   mainApp.InitCommands,
		InitContainers: make([]models.InitContainer, 0),
		Volumes:        k8sVolumes,
		Disabled:       mainApp.Replicas == 0,
	}

	if d := kg.deployment.Subsystems.K8s.GetDeployment(kg.deployment.Name); subsystems.Created(d) {
		dep.CreatedAt = d.CreatedAt
	}

	return []models.DeploymentPublic{dep}
}

func (kg *K8sGenerator) Services() []models.ServicePublic {
	mainApp := kg.deployment.GetMainApp()

	se := models.ServicePublic{
		Name:      kg.deployment.Name,
		Namespace: kg.namespace,
		Ports:     []models.Port{{Name: "http", Protocol: "tcp", Port: mainApp.InternalPort, TargetPort: mainApp.InternalPort}},
		Selector: map[string]string{
			keys.LabelDeployName: kg.deployment.Name,
		},
	}

	if k8sService := kg.deployment.Subsystems.K8s.GetService(kg.deployment.Name); subsystems.Created(k8sService) {
		se.CreatedAt = k8sService.CreatedAt
	}

	return []models.ServicePublic{se}
}

func (kg *K8sGenerator) Ingresses() []models.IngressPublic {
	var res []models.IngressPublic

	mainApp := kg.deployment.GetMainApp()
	if mainApp.Private {
		return res
	}

	// If replicas == 0, it should point to the fallback-disabled deployment
	var serviceName string
	var servicePort int
	if mainApp.Replicas == 0 {
		serviceName = config.Config.Deployment.Fallback.Disabled.Name
		servicePort = config.Config.Deployment.Port
	} else {
		serviceName = kg.deployment.Name
		servicePort = mainApp.InternalPort
	}

	tlsSecret := constants.WildcardCertSecretName
	in := models.IngressPublic{
		Name:         kg.deployment.Name,
		Namespace:    kg.namespace,
		ServiceName:  serviceName,
		ServicePort:  servicePort,
		IngressClass: config.Config.Deployment.IngressClass,
		Hosts:        []string{getExternalFQDN(kg.deployment.Name, kg.zone)},
		Placeholder:  false,
		TlsSecret:    &tlsSecret,
		CustomCert:   nil,
	}

	if k8sIngress := kg.deployment.Subsystems.K8s.GetIngress(kg.deployment.Name); subsystems.Created(k8sIngress) {
		in.CreatedAt = k8sIngress.CreatedAt
	}

	res = append(res, in)

	if mainApp.CustomDomain != nil && mainApp.CustomDomain.Status == model.CustomDomainStatusActive {
		customIn := models.IngressPublic{
			Name:         fmt.Sprintf(constants.WithCustomDomainSuffix(kg.deployment.Name)),
			Namespace:    kg.namespace,
			ServiceName:  kg.deployment.Name,
			ServicePort:  mainApp.InternalPort,
			IngressClass: config.Config.Deployment.IngressClass,
			Hosts:        []string{mainApp.CustomDomain.Domain},
			CustomCert: &models.CustomCert{
				ClusterIssuer: kg.zone.K8s.ClusterIssuer,
				CommonName:    mainApp.CustomDomain.Domain,
			},
			TlsSecret: nil,
		}

		if customK8sIngress := kg.deployment.Subsystems.K8s.GetIngress(constants.WithCustomDomainSuffix(kg.deployment.Name)); subsystems.Created(customK8sIngress) {
			customIn.CreatedAt = customK8sIngress.CreatedAt
		}

		res = append(res, customIn)
	}

	return res
}

func (kg *K8sGenerator) PVs() []models.PvPublic {
	res := make([]models.PvPublic, 0)
	if len(kg.deployment.GetMainApp().Volumes) == 0 {
		return res
	}

	volumes := kg.deployment.GetMainApp().Volumes

	for _, v := range volumes {
		res = append(res, models.PvPublic{
			Name:      deploymentPvName(kg.deployment, v.Name),
			Capacity:  fmt.Sprintf("%.1fGi", config.Config.Deployment.Resources.Limits.Storage),
			NfsServer: kg.zone.Storage.NfsServer,
			NfsPath:   path.Join(kg.zone.Storage.Paths.ParentDeployment, kg.deployment.OwnerID, "user", v.ServerPath),
			Released:  false,
		})
	}

	// Add volume to root that can be used to create storage paths
	res = append(res, models.PvPublic{
		Name:      deploymentRootPvName(kg.deployment),
		Capacity:  fmt.Sprintf("%.1fGi", config.Config.Deployment.Resources.Limits.Storage),
		NfsServer: kg.zone.Storage.NfsServer,
		NfsPath:   path.Join(kg.zone.Storage.Paths.ParentDeployment, kg.deployment.OwnerID, "user"),
		Released:  false,
	})

	for mapName, pv := range kg.deployment.Subsystems.K8s.GetPvMap() {
		idx := slices.IndexFunc(res, func(pv models.PvPublic) bool {
			return pv.Name == mapName
		})
		if idx != -1 {
			res[idx].CreatedAt = pv.CreatedAt
		}
	}

	return res
}

func (kg *K8sGenerator) PVCs() []models.PvcPublic {
	res := make([]models.PvcPublic, 0)
	if len(kg.deployment.GetMainApp().Volumes) == 0 {
		return res
	}

	volumes := kg.deployment.GetMainApp().Volumes

	for _, volume := range volumes {
		res = append(res, models.PvcPublic{
			Name:      deploymentPvcName(kg.deployment, volume.Name),
			Namespace: kg.namespace,
			Capacity:  fmt.Sprintf("%.1fGi", config.Config.Deployment.Resources.Limits.Storage),
			PvName:    deploymentPvName(kg.deployment, volume.Name),
		})
	}

	// Add PVC for root that can be used to create storage paths
	res = append(res, models.PvcPublic{
		Name:      deploymentRootPvcName(kg.deployment),
		Namespace: kg.namespace,
		Capacity:  fmt.Sprintf("%.1fGi", config.Config.Deployment.Resources.Limits.Storage),
		PvName:    deploymentRootPvName(kg.deployment),
	})

	for mapName, pvc := range kg.deployment.Subsystems.K8s.GetPvcMap() {
		idx := slices.IndexFunc(res, func(pvc models.PvcPublic) bool {
			return pvc.Name == mapName
		})
		if idx != -1 {
			res[idx].CreatedAt = pvc.CreatedAt
		}
	}

	return res
}

func (kg *K8sGenerator) Secrets() []models.SecretPublic {
	res := make([]models.SecretPublic, 0)

	if kg.deployment.Type == model.DeploymentTypeCustom {
		var imagePullSecret *models.SecretPublic

		if kg.deployment.Subsystems.Harbor.Robot.Created() && kg.deployment.Type == model.DeploymentTypeCustom {
			registry := config.Config.Registry.URL
			username := kg.deployment.Subsystems.Harbor.Robot.HarborName
			password := kg.deployment.Subsystems.Harbor.Robot.Secret

			imagePullSecret = &models.SecretPublic{
				Name:      constants.WithImagePullSecretSuffix(kg.deployment.Name),
				Namespace: kg.namespace,
				Type:      string(v1.SecretTypeDockerConfigJson),
				Data: map[string][]byte{
					v1.DockerConfigJsonKey: encodeDockerConfig(registry, username, password),
				},
			}
		}

		// if already exists, set the fields that are created in the subsystem
		if secret := kg.deployment.Subsystems.K8s.GetSecret(constants.WithImagePullSecretSuffix(kg.deployment.Name)); subsystems.Created(secret) {
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

	if secret := kg.deployment.Subsystems.K8s.GetSecret(constants.WildcardCertSecretName); subsystems.Created(secret) {
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

func (kg *K8sGenerator) HPAs() []models.HpaPublic {
	res := make([]models.HpaPublic, 0)

	mainApp := kg.deployment.GetMainApp()

	minReplicas := 1
	maxReplicas := mainApp.Replicas

	// If replicas == 0, it should point to the fallback-disabled deployment
	// This means we can delete the HPA
	if mainApp.Replicas == 0 {
		return res
	}

	hpa := models.HpaPublic{
		Name:        kg.deployment.Name,
		Namespace:   kg.namespace,
		MinReplicas: minReplicas,
		MaxReplicas: maxReplicas,
		Target: models.Target{
			Kind:       "Deployment",
			Name:       kg.deployment.Name,
			ApiVersion: "apps/v1",
		},
		CpuAverageUtilization:    config.Config.Deployment.Resources.AutoScale.CpuThreshold,
		MemoryAverageUtilization: config.Config.Deployment.Resources.AutoScale.MemoryThreshold,
	}

	if h := kg.deployment.Subsystems.K8s.GetHPA(kg.deployment.Name); subsystems.Created(h) {
		hpa.CreatedAt = h.CreatedAt
	}

	res = append(res, hpa)
	return res
}

func (kg *K8sGenerator) OneShotJobs() []models.JobPublic {
	res := make([]models.JobPublic, 0)
	if len(kg.deployment.GetMainApp().Volumes) == 0 {
		return res
	}

	// OneShot jobs generate the path in the storage server for the user

	args := []string{
		"-p",
	}
	for _, v := range kg.deployment.GetMainApp().Volumes {
		if v.ServerPath == "" {
			continue
		}

		args = append(args, path.Join("/exports", v.ServerPath))
	}

	pvcName := deploymentRootPvcName(kg.deployment)
	res = append(res, models.JobPublic{
		Name:      fmt.Sprintf("create-storage-path-%s", kg.deployment.Name),
		Namespace: kg.namespace,
		Image:     "busybox",
		Command:   []string{"/bin/mkdir"},
		Args:      args,
		Volumes: []models.Volume{
			{
				Name:      "storage",
				PvcName:   &pvcName,
				MountPath: "/exports",
				Init:      false,
			},
		},
		CreatedAt: time.Now(),
	})

	return res
}

func (kg *K8sGenerator) NetworkPolicies() []models.NetworkPolicyPublic {
	res := make([]models.NetworkPolicyPublic, 0)

	for _, egressRule := range kg.zone.NetworkPolicies {
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
			Name:        deploymentNetworkPolicyName(kg.deployment.Name, egressRule.Name),
			Namespace:   kg.namespace,
			Selector:    map[string]string{keys.LabelDeployName: kg.deployment.Name},
			EgressRules: egressRules,
			IngressRules: []models.IngressRule{
				{
					PodSelector:       map[string]string{"owner-id": kg.deployment.OwnerID},
					NamespaceSelector: map[string]string{"kubernetes.io/metadata.name": kg.namespace},
				},
				{
					NamespaceSelector: map[string]string{"kubernetes.io/metadata.name": kg.zone.K8s.IngressNamespace},
				},
			},
		}

		if npo := kg.deployment.Subsystems.K8s.GetNetworkPolicy(egressRule.Name); subsystems.Created(npo) {
			np.CreatedAt = npo.CreatedAt
		}

		res = append(res, np)
	}

	return res
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

// deploymentPvName returns the PV name for a deployment
func deploymentPvName(deployment *model.Deployment, volumeName string) string {
	return fmt.Sprintf("%s-%s", deployment.Name, makeValidK8sName(volumeName))
}

// deploymentRootPvName returns the root PV name for a deployment
func deploymentRootPvName(deployment *model.Deployment) string {
	return fmt.Sprintf("root-%s", deployment.Name)
}

// deploymentPvcName returns the PVC name for a deployment
func deploymentPvcName(deployment *model.Deployment, volumeName string) string {
	return fmt.Sprintf("%s-%s", deployment.Name, makeValidK8sName(volumeName))
}

// deploymentRootPvcName returns the root PVC name for a deployment
func deploymentRootPvcName(deployment *model.Deployment) string {
	return fmt.Sprintf("root-%s", deployment.Name)
}

// deploymentNetworkPolicyName returns the network policy name for a VM or Deployment
func deploymentNetworkPolicyName(name, egressRuleName string) string {
	return fmt.Sprintf("%s-%s", name, egressRuleName)
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

// getExternalFQDN returns the external FQDN for a deployment in a given zone
func getExternalFQDN(name string, zone *configModels.Zone) string {
	// Remove protocol:// and :port from the zone.Domains.ParentDeployment
	var fqdn = zone.Domains.ParentDeployment

	split := strings.Split(zone.Domains.ParentDeployment, "://")
	if len(split) > 1 {
		fqdn = split[1]
	}

	fqdn = strings.Split(fqdn, ":")[0]

	return fmt.Sprintf("%s.%s", name, fqdn)
}

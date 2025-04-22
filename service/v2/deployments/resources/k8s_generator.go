package resources

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	configModels "github.com/kthcloud/go-deploy/models/config"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/db/resources/team_repo"
	"github.com/kthcloud/go-deploy/pkg/db/resources/user_repo"
	"github.com/kthcloud/go-deploy/pkg/subsystems"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/keys"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/models"
	"github.com/kthcloud/go-deploy/service/constants"
	"github.com/kthcloud/go-deploy/service/generators"
	"github.com/kthcloud/go-deploy/utils"
	v1 "k8s.io/api/core/v1"
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
		if env.Name == "PORT" || env.Name == "INTERNAL_PORTS" {
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

	portsStr := make([]string, len(mainApp.InternalPorts))
	for i, port := range mainApp.InternalPorts {
		portsStr[i] = strconv.Itoa(port)
	}

	k8sEnvs = append(k8sEnvs, models.EnvVar{
		Name:  "INTERNAL_PORTS",
		Value: strings.Join(portsStr, ","),
	})

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

	res := make([]models.DeploymentPublic, 0)

	dep := models.DeploymentPublic{
		Name:             kg.deployment.Name,
		Namespace:        kg.namespace,
		Labels:           map[string]string{"owner-id": kg.deployment.OwnerID},
		Image:            mainApp.Image,
		ImagePullSecrets: imagePullSecrets,
		EnvVars:          k8sEnvs,
		Resources: models.Resources{
			Limits: models.Limits{
				CPU:    formatCpuString(mainApp.CpuCores),
				Memory: fmt.Sprintf("%dMi", int(mainApp.RAM*1000)),
			},
			Requests: models.Requests{
				CPU:    formatCpuString(math.Min(config.Config.Deployment.Resources.Requests.CPU, mainApp.CpuCores)),
				Memory: fmt.Sprintf("%dMi", int(math.Min(config.Config.Deployment.Resources.Requests.RAM, mainApp.RAM)*1000)),
			},
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

	res = append(res, dep)

	if mainApp.Visibility == model.VisibilityAuth && mainApp.Replicas > 0 {

		generateAuthProxy := func() (*models.DeploymentPublic, error) {
			// Auth proxy

			//// Find users that should be able to access the resource
			teamIDs, err := team_repo.New().WithResourceID(kg.deployment.ID).ListIDs()
			if err != nil {
				return nil, err
			}

			memberIDs, err := team_repo.New().ListMemberIDs(teamIDs...)
			if err != nil {
				return nil, err
			}

			userEmailMap, err := user_repo.New().ListEmails(memberIDs...)
			if err != nil {
				return nil, err
			}

			// Ensure owner is included
			if ownerEmail, err := user_repo.New().GetEmail(kg.deployment.OwnerID); err == nil {
				userEmailMap[kg.deployment.OwnerID] = ownerEmail
			}

			command := "mkdir -p /mnt/config && echo \""
			for _, email := range userEmailMap {
				command += email + "\n"
			}
			command += "\" > /mnt/config/authenticated-emails-list"

			var redirectURL string
			if mainApp.CustomDomain != nil {
				redirectURL = fmt.Sprintf("https://%s/oauth2/callback", mainApp.CustomDomain.Domain)
			} else {
				redirectURL = fmt.Sprintf("https://%s.%s/oauth2/callback", kg.deployment.Name, kg.zone.Domains.ParentDeployment)
			}

			oauthProxy := models.DeploymentPublic{
				Name:             authProxyName(kg.deployment.Name),
				Namespace:        kg.namespace,
				Labels:           map[string]string{"owner-id": kg.deployment.OwnerID},
				Image:            "quay.io/oauth2-proxy/oauth2-proxy:latest",
				ImagePullSecrets: make([]string, 0),
				EnvVars:          make([]models.EnvVar, 0),
				Resources: models.Resources{
					Limits: models.Limits{
						CPU:    formatCpuString(config.Config.Deployment.Resources.Limits.CPU),
						Memory: fmt.Sprintf("%dMi", int(config.Config.Deployment.Resources.Limits.RAM*1000)),
					},
					Requests: models.Requests{
						CPU:    formatCpuString(config.Config.Deployment.Resources.Requests.CPU),
						Memory: fmt.Sprintf("%dMi", int(config.Config.Deployment.Resources.Requests.RAM*1000)),
					},
				},
				Command: make([]string, 0),
				Args: []string{
					"--http-address=0.0.0.0:4180",
					"--reverse-proxy=true",
					"--provider=oidc",
					"--redirect-url=" + redirectURL,
					"--oidc-issuer-url=" + config.Config.Keycloak.Url + "/realms/" + config.Config.Keycloak.Realm,
					"--cookie-expire=168h",
					"--cookie-refresh=1h",
					"--pass-authorization-header=true",
					"--scope=openid email",
					"--upstream=" + fmt.Sprintf("http://%s:%d", kg.deployment.Name, mainApp.InternalPort),
					"--client-id=" + config.Config.Keycloak.UserClient.ClientID,
					"--client-secret=" + config.Config.Keycloak.UserClient.ClientSecret,
					"--cookie-secret=qHKgjlAFQBZOnGcdH5jIKV0Auzx5r8jzZenxhJnlZJg=",
					"--cookie-secure=true",
					"--ssl-insecure-skip-verify=true",
					"--insecure-oidc-allow-unverified-email=true",
					"--skip-provider-button=true",
					"--pass-authorization-header=true",
					"--ssl-upstream-insecure-skip-verify=true",
					"--code-challenge-method=S256",
					"--authenticated-emails-file=/mnt/authenticated-emails-list",
				},
				InitCommands: make([]string, 0),
				InitContainers: []models.InitContainer{{
					Name:    "oauth-proxy-config-init",
					Image:   "busybox",
					Command: []string{"sh", "-c", command},
					Args:    nil,
				}},
				Volumes: []models.Volume{
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
				},
			}

			if op := kg.deployment.Subsystems.K8s.GetDeployment(authProxyName(kg.deployment.Name)); subsystems.Created(op) {
				oauthProxy.CreatedAt = op.CreatedAt
			}

			return &oauthProxy, nil
		}

		authProxy, err := generateAuthProxy()
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to generate auth proxy for deployment %s. details: %w", kg.deployment.Name, err))
		} else {
			res = append(res, *authProxy)
		}
	}

	return res
}

func (kg *K8sGenerator) Services() []models.ServicePublic {
	mainApp := kg.deployment.GetMainApp()

	// If replicas == 0, it should not create a service
	// If visibility == auth, it should create both a service for the deployment and the auth proxy

	if mainApp.Replicas == 0 {
		return make([]models.ServicePublic, 0)
	}

	res := make([]models.ServicePublic, 0)

	// Add the base http port
	ports := []models.Port{
		{
			Name:       "http",
			Protocol:   "tcp",
			Port:       mainApp.InternalPort,
			TargetPort: mainApp.InternalPort,
		},
	}

	// add all internalPorts to expose to the with the service
	for _, p := range mainApp.InternalPorts {
		if p == mainApp.InternalPort || p == 0 {
			continue
		}

		ports = append(ports, models.Port{
			Name:       fmt.Sprintf("port-%d", p),
			Protocol:   "tcp",
			Port:       p,
			TargetPort: p,
		})
	}

	se := models.ServicePublic{
		Name:      kg.deployment.Name,
		Namespace: kg.namespace,
		Ports:     ports,
		Selector: map[string]string{
			keys.LabelDeployName: kg.deployment.Name,
		},
	}

	if k8sService := kg.deployment.Subsystems.K8s.GetService(kg.deployment.Name); subsystems.Created(k8sService) {
		se.CreatedAt = k8sService.CreatedAt
	}

	res = append(res, se)

	if mainApp.Visibility == model.VisibilityAuth {
		authSe := models.ServicePublic{
			Name:      authProxyName(kg.deployment.Name),
			Namespace: kg.namespace,
			Ports:     []models.Port{{Name: "http", Protocol: "tcp", Port: 80, TargetPort: 4180}},
			Selector: map[string]string{
				keys.LabelDeployName: authProxyName(kg.deployment.Name),
			},
		}

		if k8sService := kg.deployment.Subsystems.K8s.GetService(authProxyName(kg.deployment.Name)); subsystems.Created(k8sService) {
			authSe.CreatedAt = k8sService.CreatedAt
		}

		res = append(res, authSe)
	}

	return res
}

func (kg *K8sGenerator) Ingresses() []models.IngressPublic {
	var res []models.IngressPublic

	mainApp := kg.deployment.GetMainApp()
	if mainApp.Visibility == model.VisibilityPrivate {
		return res
	}

	var serviceName string
	var servicePort int

	// If replicas == 0, it should point to the fallback-disabled deployment
	// If visibility == auth, it should point to the auth proxy
	// Otherwise, it should point to the deployment itself

	if mainApp.Replicas == 0 {
		serviceName = config.Config.Deployment.Fallback.Disabled.Name
		servicePort = config.Config.Deployment.Port
	} else if mainApp.Visibility == model.VisibilityAuth {
		serviceName = authProxyName(kg.deployment.Name)
		servicePort = 4180
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
			Name:         fmt.Sprint(constants.WithCustomDomainSuffix(kg.deployment.Name)),
			Namespace:    kg.namespace,
			ServiceName:  serviceName,
			ServicePort:  servicePort,
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
			Capacity:  fmt.Sprintf("%dGi", config.Config.Deployment.Resources.Limits.Storage),
			NfsServer: kg.zone.Storage.NfsServer,
			NfsPath:   path.Join(kg.zone.Storage.Paths.ParentDeployment, kg.deployment.OwnerID, "user", v.ServerPath),
			Released:  false,
		})
	}

	// Add volume to root that can be used to create storage paths
	res = append(res, models.PvPublic{
		Name:      deploymentRootPvName(kg.deployment),
		Capacity:  fmt.Sprintf("%dGi", config.Config.Deployment.Resources.Limits.Storage),
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
			Capacity:  fmt.Sprintf("%dGi", config.Config.Deployment.Resources.Limits.Storage),
			PvName:    deploymentPvName(kg.deployment, volume.Name),
		})
	}

	// Add PVC for root that can be used to create storage paths
	res = append(res, models.PvcPublic{
		Name:      deploymentRootPvcName(kg.deployment),
		Namespace: kg.namespace,
		Capacity:  fmt.Sprintf("%dGi", config.Config.Deployment.Resources.Limits.Storage),
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

// authProxyName returns the name of the auth proxy for a deployment
func authProxyName(name string) string {
	return fmt.Sprintf("%s-auth-proxy", name)
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

// formatCpuString formats the CPU string.
// It ensures the same is returned by K8s after creation.
func formatCpuString(cpus float64) string {
	// Round to one decimal, e.g. 0.12 -> 0.1, 0.16 -> 0.2
	oneDec := math.Round(cpus*10) / 10

	// If whole number, return as int, e.g. 1.0 -> 1
	if oneDec == float64(int(oneDec)) {
		return fmt.Sprintf("%d", int(oneDec))
	}

	// Otherwise, return as milli CPU, e.g. 0.1 -> 100m
	return fmt.Sprintf("%dm", int(oneDec*1000))
}

package resources

import (
	"fmt"
	configModels "go-deploy/models/config"
	"go-deploy/models/model"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/user_repo"
	"go-deploy/pkg/subsystems"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/pkg/subsystems/k8s/keys"
	"go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service/constants"
	"go-deploy/service/generators"
	"go-deploy/utils"
	v1 "k8s.io/api/core/v1"
	"path"
	"slices"
)

type K8sGenerator struct {
	generators.K8sGeneratorBase

	namespace string
	client    *k8s.Client

	sm   *model.SM
	zone *configModels.Zone
}

func K8s(sm *model.SM, zone *configModels.Zone, client *k8s.Client, namespace string) *K8sGenerator {
	return &K8sGenerator{
		namespace: namespace,
		client:    client,
		sm:        sm,
		zone:      zone,
	}
}

func (kg *K8sGenerator) Namespace() *models.NamespacePublic {
	ns := models.NamespacePublic{
		Name: kg.namespace,
	}

	if n := &kg.sm.Subsystems.K8s.Namespace; subsystems.Created(n) {
		ns.CreatedAt = n.CreatedAt
	}

	return &ns
}

func (kg *K8sGenerator) Deployments() []models.DeploymentPublic {
	var res []models.DeploymentPublic

	initVolumes, stdVolume := smVolumes(kg.sm.OwnerID)
	allVolumes := append(initVolumes, stdVolume...)

	k8sVolumes := make([]models.Volume, len(allVolumes))
	for i, volume := range allVolumes {
		pvcName := smPvcName(kg.sm.OwnerID, volume.Name)
		k8sVolumes[i] = models.Volume{
			Name:      smPvName(kg.sm.OwnerID, volume.Name),
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
		Name:             smName(kg.sm.OwnerID),
		Namespace:        kg.namespace,
		Labels:           map[string]string{"owner-id": kg.sm.OwnerID},
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

	if fb := kg.sm.Subsystems.K8s.GetDeployment(smName(kg.sm.OwnerID)); subsystems.Created(fb) {
		filebrowser.CreatedAt = fb.CreatedAt
	}

	res = append(res, filebrowser)

	// Oauth2-proxy
	user, err := user_repo.New().GetByID(kg.sm.OwnerID)
	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("failed to get user by id when creating oauth proxy deployment public. details: %w", err))
		return nil
	}

	if user == nil {
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
	redirectURL := fmt.Sprintf("https://%s.%s/oauth2/callback", kg.sm.OwnerID, kg.zone.Domains.ParentSM)
	upstream := "http://" + smName(kg.sm.OwnerID) + ":80"

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
		Name:             smAuthName(kg.sm.OwnerID),
		Namespace:        kg.namespace,
		Labels:           map[string]string{"owner-id": kg.sm.OwnerID},
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

	if op := kg.sm.Subsystems.K8s.GetDeployment(smAuthName(kg.sm.OwnerID)); subsystems.Created(op) {
		oauthProxy.CreatedAt = op.CreatedAt
	}

	res = append(res, oauthProxy)

	return res
}

func (kg *K8sGenerator) Services() []models.ServicePublic {
	res := make([]models.ServicePublic, 0)

	// Filebrowser
	filebrowser := models.ServicePublic{
		Name:      smName(kg.sm.OwnerID),
		Namespace: kg.namespace,
		Ports:     []models.Port{{Name: "http", Protocol: "tcp", Port: 80, TargetPort: 80}},
		Selector: map[string]string{
			keys.LabelDeployName: smName(kg.sm.OwnerID),
		},
	}

	if fb := kg.sm.Subsystems.K8s.GetService(smName(kg.sm.OwnerID)); subsystems.Created(fb) {
		filebrowser.CreatedAt = fb.CreatedAt
	}

	res = append(res, filebrowser)

	// Oauth2-proxy
	oauthProxy := models.ServicePublic{
		Name:      smAuthName(kg.sm.OwnerID),
		Namespace: kg.namespace,
		Ports:     []models.Port{{Name: "http", Protocol: "tcp", Port: 4180, TargetPort: 4180}},
		Selector: map[string]string{
			keys.LabelDeployName: smAuthName(kg.sm.OwnerID),
		},
	}

	if op := kg.sm.Subsystems.K8s.GetService(smAuthName(kg.sm.OwnerID)); subsystems.Created(op) {
		oauthProxy.CreatedAt = op.CreatedAt
	}

	res = append(res, oauthProxy)

	return res
}

func (kg *K8sGenerator) Ingresses() []models.IngressPublic {
	res := make([]models.IngressPublic, 0)

	tlsSecret := constants.WildcardCertSecretName

	ingress := models.IngressPublic{
		Name:         smName(kg.sm.OwnerID),
		Namespace:    kg.namespace,
		ServiceName:  smAuthName(kg.sm.OwnerID),
		ServicePort:  4180,
		IngressClass: config.Config.Deployment.IngressClass,
		Hosts:        []string{storageExternalFQDN(kg.sm.OwnerID, kg.zone)},
		TlsSecret:    &tlsSecret,
	}

	if i := kg.sm.Subsystems.K8s.GetIngress(smName(kg.sm.OwnerID)); subsystems.Created(i) {
		ingress.CreatedAt = i.CreatedAt
	}

	res = append(res, ingress)
	return res
}

func (kg *K8sGenerator) PVs() []models.PvPublic {
	res := make([]models.PvPublic, 0)

	initVolumes, volumes := smVolumes(kg.sm.OwnerID)
	allVolumes := append(initVolumes, volumes...)

	for _, v := range allVolumes {
		res = append(res, models.PvPublic{
			Name:      smPvName(kg.sm.OwnerID, v.Name),
			Capacity:  config.Config.Deployment.Resources.Limits.Storage,
			NfsServer: kg.zone.Storage.NfsServer,
			NfsPath:   path.Join(kg.zone.Storage.Paths.ParentDeployment, v.ServerPath),
			Released:  false,
		})
	}

	for mapName, pv := range kg.sm.Subsystems.K8s.GetPvMap() {
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

	initVolumes, volumes := smVolumes(kg.sm.OwnerID)
	allVolumes := append(initVolumes, volumes...)

	for _, volume := range allVolumes {
		res = append(res, models.PvcPublic{
			Name:      smPvcName(kg.sm.OwnerID, volume.Name),
			Namespace: kg.namespace,
			Capacity:  config.Config.Deployment.Resources.Limits.Storage,
			PvName:    smPvName(kg.sm.OwnerID, volume.Name),
		})
	}

	for mapName, pvc := range kg.sm.Subsystems.K8s.GetPvcMap() {
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

	// Wildcard certificate
	/// Swap namespaces temporarily
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

	if secret := kg.sm.Subsystems.K8s.GetSecret(constants.WildcardCertSecretName); subsystems.Created(secret) {
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

func (kg *K8sGenerator) OneShotJobs() []models.JobPublic {
	res := make([]models.JobPublic, 0)

	// These are assumed to be one-shot jobs
	initVolumes, _ := smVolumes(kg.sm.OwnerID)
	k8sVolumes := make([]models.Volume, len(initVolumes))
	for i, initVolume := range initVolumes {
		pvcName := smPvcName(kg.sm.OwnerID, initVolume.Name)
		k8sVolumes[i] = models.Volume{
			Name:      smPvName(kg.sm.OwnerID, initVolume.Name),
			PvcName:   &pvcName,
			MountPath: initVolume.AppPath,
			Init:      initVolume.Init,
		}
	}

	for _, job := range smJobs(kg.sm.OwnerID) {
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

// smVolumes returns the init and standard volumes for a storage manager
func smVolumes(ownerID string) ([]model.SmVolume, []model.SmVolume) {
	initVolumes := []model.SmVolume{
		{
			Name:       "init",
			Init:       false,
			AppPath:    "/exports",
			ServerPath: "",
		},
	}

	volumes := []model.SmVolume{
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

// smJobs returns the init jobs for a storage manager
func smJobs(userID string) []model.SmJob {
	return []model.SmJob{
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

// smName returns the name for a storage manager
func smName(userID string) string {
	return fmt.Sprintf("%s-%s", constants.SmAppName, userID)
}

// smAuthName returns the name for a storage manager auth proxy
func smAuthName(userID string) string {
	return fmt.Sprintf("%s-%s", constants.SmAppNameAuth, userID)
}

// smPvcName returns the PVC name for a storage manager
func smPvcName(ownerID, volumeName string) string {
	return fmt.Sprintf("sm-%s-%s", volumeName, ownerID)
}

// smPvName returns the PV name for a storage manager
func smPvName(ownerID, volumeName string) string {
	return fmt.Sprintf("sm-%s-%s", volumeName, ownerID)
}

// storageExternalFQDN returns the external FQDN for a storage manager in a given zone
func storageExternalFQDN(name string, zone *configModels.Zone) string {
	return fmt.Sprintf("%s.%s", name, zone.Domains.ParentSM)
}

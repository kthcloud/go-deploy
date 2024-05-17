package model

import (
	"fmt"
	"go-deploy/dto/v1/body"
	"go-deploy/pkg/log"
	"go-deploy/utils"
	"golang.org/x/net/idna"
	"strconv"
)

// ToDTO converts a Deployment to a body.DeploymentRead DTO.
func (deployment *Deployment) ToDTO(smURL *string, externalPort *int, teams []string) body.DeploymentRead {
	app := deployment.GetMainApp()
	if app == nil {
		log.Println("Main app not found in deployment", deployment.ID)
		app = &App{}
	}

	envs := make([]body.Env, len(app.Envs))

	envs = append(envs, body.Env{
		Name:  "PORT",
		Value: fmt.Sprintf("%d", app.InternalPort),
	})

	for i, env := range app.Envs {
		envs[i] = body.Env{
			Name:  env.Name,
			Value: env.Value,
		}
	}

	volumes := make([]body.Volume, len(app.Volumes))
	for i, volume := range app.Volumes {
		volumes[i] = body.Volume{
			Name:       volume.Name,
			AppPath:    volume.AppPath,
			ServerPath: volume.ServerPath,
		}
	}

	if app.InitCommands == nil {
		app.InitCommands = make([]string, 0)
	}

	if app.Args == nil {
		app.Args = make([]string, 0)
	}

	var pingResult *int
	if app.PingResult != 0 {
		pingResult = &app.PingResult
	}

	var image *string
	if deployment.Type == DeploymentTypePrebuilt {
		image = &app.Image
	}

	var healthCheckPath *string
	if app.PingPath != "" {
		healthCheckPath = &app.PingPath
	}

	var customDomain *string
	if app.CustomDomain != nil {
		customDomain = &app.CustomDomain.Domain
	}

	var customDomainSecret *string
	if app.CustomDomain != nil {
		customDomainSecret = &app.CustomDomain.Secret
	}

	var customDomainStatus *string
	if app.CustomDomain != nil {
		customDomainStatus = &app.CustomDomain.Status
	}

	var deploymentError *string
	if deployment.Error != nil {
		deploymentError = &deployment.Error.Description
	}

	var status string
	if deploymentError == nil {
		status = deployment.Status
	} else {
		status = deployment.Error.Reason
	}

	var replicaStatus *body.ReplicaStatus
	if app.ReplicaStatus != nil {
		replicaStatus = &body.ReplicaStatus{
			DesiredReplicas:     app.ReplicaStatus.DesiredReplicas,
			ReadyReplicas:       app.ReplicaStatus.ReadyReplicas,
			AvailableReplicas:   app.ReplicaStatus.AvailableReplicas,
			UnavailableReplicas: app.ReplicaStatus.UnavailableReplicas,
		}
	}

	return body.DeploymentRead{
		ID:      deployment.ID,
		Name:    deployment.Name,
		Type:    deployment.Type,
		OwnerID: deployment.OwnerID,
		Zone:    deployment.Zone,

		CreatedAt:   deployment.CreatedAt,
		UpdatedAt:   utils.NonZeroOrNil(deployment.UpdatedAt),
		RepairedAt:  utils.NonZeroOrNil(deployment.RepairedAt),
		RestartedAt: utils.NonZeroOrNil(deployment.RestartedAt),
		AccessedAt:  deployment.AccessedAt,

		CpuCores: app.CpuCores,
		RAM:      app.RAM,
		Replicas: app.Replicas,

		URL:             deployment.GetURL(externalPort),
		Envs:            envs,
		Volumes:         volumes,
		InitCommands:    app.InitCommands,
		Args:            app.Args,
		Private:         app.Private,
		InternalPort:    app.InternalPort,
		Image:           image,
		HealthCheckPath: healthCheckPath,

		CustomDomain:       customDomain,
		CustomDomainURL:    deployment.GetCustomDomainURL(),
		CustomDomainSecret: customDomainSecret,
		CustomDomainStatus: customDomainStatus,

		Status:        status,
		Error:         deploymentError,
		ReplicaStatus: replicaStatus,
		PingResult:    pingResult,

		Integrations: make([]string, 0),

		StorageURL: smURL,
		Teams:      teams,
	}
}

// FromDTO converts body.DeploymentCreate DTO to DeploymentCreateParams.
func (p *DeploymentCreateParams) FromDTO(dto *body.DeploymentCreate, fallbackZone, fallbackImage string, fallbackPort int) {
	p.Name = dto.Name

	if dto.Image == nil {
		p.Image = fallbackImage
		p.Type = DeploymentTypeCustom
	} else {
		p.Image = *dto.Image
		p.Type = DeploymentTypePrebuilt
	}

	if dto.CpuCores != nil {
		p.CpuCores = *dto.CpuCores
	}

	if dto.RAM != nil {
		p.RAM = *dto.RAM
	}

	p.Private = dto.Private
	p.Envs = make([]DeploymentEnv, 0)
	for _, env := range dto.Envs {
		if env.Name == "PORT" {
			port, _ := strconv.Atoi(env.Value)
			p.InternalPort = port
			continue
		}

		p.Envs = append(p.Envs, DeploymentEnv{
			Name:  env.Name,
			Value: env.Value,
		})
	}

	// if user didn't specify $PORT
	if p.InternalPort == 0 {
		p.InternalPort = fallbackPort
	}

	p.Volumes = make([]DeploymentVolume, len(dto.Volumes))
	for i, volume := range dto.Volumes {
		p.Volumes[i] = DeploymentVolume{
			Name:       volume.Name,
			AppPath:    volume.AppPath,
			ServerPath: volume.ServerPath,
			Init:       false,
		}
	}
	p.InitCommands = dto.InitCommands
	p.Args = dto.Args

	if dto.HealthCheckPath == nil {
		p.PingPath = "/healthz"
	} else {
		p.PingPath = *dto.HealthCheckPath
	}

	if dto.CustomDomain != nil {
		if punyEncoded, err := idna.New().ToASCII(*dto.CustomDomain); err == nil {
			p.CustomDomain = &punyEncoded
		} else {
			utils.PrettyPrintError(fmt.Errorf("failed to puny encode domain %s when creating create params details: %w", *dto.CustomDomain, err))
		}
	}

	p.Replicas = 1
	if dto.Replicas != nil {
		p.Replicas = *dto.Replicas
	}

	if dto.Zone != nil {
		p.Zone = *dto.Zone
	} else {
		p.Zone = fallbackZone
	}
}

// FromDTO converts body.DeploymentUpdate DTO to DeploymentUpdateParams.
func (p *DeploymentUpdateParams) FromDTO(dto *body.DeploymentUpdate, deploymentType string) {
	if dto.Envs != nil {
		envs := make([]DeploymentEnv, 0)
		for _, env := range *dto.Envs {
			if env.Name == "PORT" {
				port, _ := strconv.Atoi(env.Value)
				p.InternalPort = &port
				continue
			}

			envs = append(envs, DeploymentEnv{
				Name:  env.Name,
				Value: env.Value,
			})
		}
		p.Envs = &envs
	}

	if dto.Volumes != nil {
		volumes := make([]DeploymentVolume, len(*dto.Volumes))
		for i, volume := range *dto.Volumes {
			volumes[i] = DeploymentVolume{
				Name:       volume.Name,
				AppPath:    volume.AppPath,
				ServerPath: volume.ServerPath,
			}
		}
		p.Volumes = &volumes
	}

	// Convert custom domain to puny encoded
	if dto.CustomDomain != nil {
		if punyEncoded, err := idna.New().ToASCII(*dto.CustomDomain); err == nil {
			p.CustomDomain = &punyEncoded
		} else {
			utils.PrettyPrintError(fmt.Errorf("failed to puny encode domain %s when creating update params details: %w", *dto.CustomDomain, err))
		}
	}

	// Only allow image updates for prebuilt deployments
	if deploymentType == DeploymentTypePrebuilt {
		p.Image = dto.Image
	}

	p.Name = dto.Name
	p.CpuCores = dto.CpuCores
	p.RAM = dto.RAM
	p.Private = dto.Private
	p.InitCommands = dto.InitCommands
	p.Args = dto.Args
	p.PingPath = dto.HealthCheckPath
	p.Replicas = dto.Replicas
}

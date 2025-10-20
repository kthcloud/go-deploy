package model

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/utils"
	"golang.org/x/net/idna"
)

// ToDTO converts a Deployment to a body.DeploymentRead DTO.
func (deployment *Deployment) ToDTO(smURL *string, externalPort *int, teams []string) body.DeploymentRead {
	app := deployment.GetMainApp()
	if app == nil {
		log.Println("Main app not found in deployment", deployment.ID)
		app = &App{}
	}

	envs := make([]body.Env, len(app.Envs))

	portIndex := -1
	internalPortIndex := -1
	for i, env := range app.Envs {
		if env.Name == "PORT" {
			portIndex = i
			continue
		} else if env.Name == "INTERNAL_PORTS" {
			internalPortIndex = i
			continue
		}
		envs[i] = body.Env{
			Name:  env.Name,
			Value: env.Value,
		}
	}

	if portIndex == -1 {
		envs = append(envs, body.Env{
			Name:  "PORT",
			Value: fmt.Sprintf("%d", app.InternalPort),
		})
	} else {
		envs[portIndex] = body.Env{
			Name:  "PORT",
			Value: fmt.Sprintf("%d", app.InternalPort),
		}
	}
	if internalPortIndex == -1 {
		if len(app.InternalPorts) > 0 {
			portsStr := make([]string, len(app.InternalPorts))
			for i, port := range app.InternalPorts {
				portsStr[i] = fmt.Sprintf("%d", port)
			}

			envs = append(envs, body.Env{
				Name:  "INTERNAL_PORTS",
				Value: strings.Join(portsStr, ","),
			})
		}
	} else if len(app.InternalPorts) > 0 {
		portsStr := make([]string, len(app.InternalPorts))
		for i, port := range app.InternalPorts {
			portsStr[i] = fmt.Sprintf("%d", port)
		}
		envs[internalPortIndex] = body.Env{
			Name:  "INTERNAL_PORTS",
			Value: strings.Join(portsStr, ","),
		}
	} else {
		envs = slices.Delete(envs, internalPortIndex, internalPortIndex+1)
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

	var customDomain *body.CustomDomainRead
	if app.CustomDomain != nil {
		extPortStr := ""
		if externalPort != nil && *externalPort != 443 {
			extPortStr = fmt.Sprintf(":%d", *externalPort)
		}
		customDomain = &body.CustomDomainRead{
			Domain: app.CustomDomain.Domain,
			URL:    fmt.Sprintf("https://%s%s", app.CustomDomain.Domain, extPortStr),
			Status: app.CustomDomain.Status,
			Secret: app.CustomDomain.Secret,
		}
	}

	var gpus []body.DeploymentGPU = make([]body.DeploymentGPU, 0, len(app.GPUs))
	for _, gpu := range app.GPUs {
		dto := body.DeploymentGPU{
			Name: gpu.Name,
		}
		if gpu.TemplateName != nil {
			dto.TemplateName = gpu.TemplateName
		} else if gpu.ClaimName != nil {
			dto.ClaimName = gpu.ClaimName
		} else {
			continue
		}

		gpus = append(gpus, dto)
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

		URL: deployment.GetURL(externalPort),
		Specs: body.DeploymentSpecs{
			CpuCores: app.CpuCores,
			RAM:      app.RAM,
			Replicas: app.Replicas,
			GPUs:     gpus,
		},

		Envs:            envs,
		Volumes:         volumes,
		InitCommands:    app.InitCommands,
		Args:            app.Args,
		InternalPort:    app.InternalPort,
		InternalPorts:   app.InternalPorts,
		Image:           image,
		HealthCheckPath: healthCheckPath,
		CustomDomain:    customDomain,
		Visibility:      app.Visibility,

		NeverStale: deployment.NeverStale,

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

	p.Envs = make([]DeploymentEnv, 0)
	for _, env := range dto.Envs {
		if env.Name == "PORT" {
			port, _ := strconv.Atoi(env.Value)
			p.InternalPort = port
			continue
		}
		if env.Name == "INTERNAL_PORTS" {
			portsStr := strings.Split(env.Value, ",")
			var internalPorts []int
			for _, prt := range portsStr {
				prt = strings.TrimSpace(prt)
				if port, err := strconv.Atoi(prt); err == nil {
					internalPorts = append(internalPorts, port)
				}
			}
			p.InternalPorts = internalPorts
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

	p.GPUs = make([]DeploymentGPU, 0, len(dto.GPUs))
	for _, gpu := range dto.GPUs {
		gpuM := DeploymentGPU{
			Name: gpu.Name,
		}
		if gpu.TemplateName != nil {
			gpuM.TemplateName = utils.StrPtr(*gpu.TemplateName)
		} else if gpu.ClaimName != nil {
			gpuM.ClaimName = utils.StrPtr(*gpu.ClaimName)
		} else {
			continue
		}
		p.GPUs = append(p.GPUs, gpuM)
	}

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

	if dto.Visibility == "" {
		p.Visibility = VisibilityPublic
	} else {
		p.Visibility = dto.Visibility
	}

	p.NeverStale = dto.NeverStale
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

			if env.Name == "INTERNAL_PORTS" {
				portsStr := strings.Split(env.Value, ",")
				var internalPorts []int
				for _, prt := range portsStr {
					prt = strings.TrimSpace(prt)
					if port, err := strconv.Atoi(prt); err == nil {
						internalPorts = append(internalPorts, port)
					}
				}
				p.InternalPorts = &internalPorts
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

	if dto.GPUs != nil {
		gpus := make([]DeploymentGPU, 0, len(*dto.GPUs))
		for _, gpu := range *dto.GPUs {
			var gpuM DeploymentGPU = DeploymentGPU{
				Name: gpu.Name,
			}

			if gpu.TemplateName != nil {
				gpuM.TemplateName = utils.StrPtr(*gpu.TemplateName)
			} else if gpu.ClaimName != nil {
				gpuM.ClaimName = utils.StrPtr(*gpu.ClaimName)
			} else {
				// One of them needs to be set
				// TODO: add validation so user can get feedback
				continue
			}

			gpus = append(gpus, gpuM)
		}
		p.GPUs = &gpus
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
	p.InitCommands = dto.InitCommands
	p.Args = dto.Args
	p.PingPath = dto.HealthCheckPath
	p.Replicas = dto.Replicas
	p.Visibility = dto.Visibility
	p.NeverStale = dto.NeverStale
}

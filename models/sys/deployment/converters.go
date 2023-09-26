package deployment

import (
	"fmt"
	"go-deploy/models/dto/body"
	"log"
	"strconv"
)

func (deployment *Deployment) ToDTO(storageManagerURL *string) body.DeploymentRead {
	app := deployment.GetMainApp()
	if app == nil {
		log.Println("main app not found in deployment", deployment.ID)
		app = &App{}
	}

	if app.Envs == nil {
		app.Envs = []Env{}
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

	integrations := make([]string, 0)
	if deployment.Subsystems.GitHub.Created() {
		integrations = append(integrations, "github")
	}

	if app.InitCommands == nil {
		app.InitCommands = make([]string, 0)
	}

	var pingResult *int
	if app.PingResult != 0 {
		pingResult = &app.PingResult
	}

	var image *string
	if deployment.Type == TypePrebuilt {
		image = &app.Image
	}

	return body.DeploymentRead{
		ID:      deployment.ID,
		Name:    deployment.Name,
		Type:    deployment.Type,
		OwnerID: deployment.OwnerID,
		Zone:    deployment.Zone,

		URL:          deployment.GetURL(),
		Envs:         envs,
		Volumes:      volumes,
		InitCommands: app.InitCommands,
		Private:      app.Private,
		InternalPort: app.InternalPort,
		Image:        image,

		Status:     deployment.StatusMessage,
		PingResult: pingResult,

		Integrations: integrations,

		StorageURL: storageManagerURL,
	}
}

func (g *GitHubRepository) ToDTO() body.GitHubRepository {
	return body.GitHubRepository{
		ID:   g.ID,
		Name: g.Name,
	}
}

func (p *UpdateParams) FromDTO(dto *body.DeploymentUpdate, deploymentType string) {
	if dto.Envs != nil {
		envs := make([]Env, 0)
		for _, env := range *dto.Envs {
			if env.Name == "PORT" {
				port, _ := strconv.Atoi(env.Value)
				p.InternalPort = &port
				continue
			}

			envs = append(envs, Env{
				Name:  env.Name,
				Value: env.Value,
			})
		}
		p.Envs = &envs
	}

	if dto.Volumes != nil {
		volumes := make([]Volume, len(*dto.Volumes))
		for i, volume := range *dto.Volumes {
			volumes[i] = Volume{
				Name:       volume.Name,
				AppPath:    volume.AppPath,
				ServerPath: volume.ServerPath,
			}
		}
		p.Volumes = &volumes
	}

	p.Private = dto.Private
	p.ExtraDomains = dto.ExtraDomains
	p.InitCommands = dto.InitCommands

	if deploymentType == TypePrebuilt {
		p.Image = dto.Image
	}
}

func (p *CreateParams) FromDTO(dto *body.DeploymentCreate, fallbackZone, fallbackImage string, fallbackPort int) {
	p.Name = dto.Name

	if dto.Image == nil {
		p.Image = fallbackImage
		p.Type = TypeCustom
	} else {
		p.Image = *dto.Image
		p.Type = TypePrebuilt
	}

	p.Private = dto.Private
	p.Envs = make([]Env, 0)
	for _, env := range dto.Envs {
		if env.Name == "PORT" {
			port, _ := strconv.Atoi(env.Value)
			p.InternalPort = port
			continue
		}

		p.Envs = append(p.Envs, Env{
			Name:  env.Name,
			Value: env.Value,
		})
	}

	// if user didn't specify $PORT
	if p.InternalPort == 0 {
		p.InternalPort = fallbackPort
	}

	p.Volumes = make([]Volume, len(dto.Volumes))
	for i, volume := range dto.Volumes {
		p.Volumes[i] = Volume{
			Name:       volume.Name,
			AppPath:    volume.AppPath,
			ServerPath: volume.ServerPath,
			Init:       false,
		}
	}
	p.InitCommands = dto.InitCommands
	p.ExtraDomains = dto.ExtraDomains

	// only allow GitHub on non-prebuilt deployments
	if p.Type == TypeCustom && dto.GitHub != nil {
		p.GitHub = &GitHubCreateParams{
			Token:        dto.GitHub.Token,
			RepositoryID: dto.GitHub.RepositoryID,
		}
	}

	if dto.Zone != nil {
		p.Zone = *dto.Zone
	} else {
		p.Zone = fallbackZone
	}
}

func (p *BuildParams) FromDTO(dto *body.DeploymentBuild) {
	p.Tag = dto.Tag
	p.Branch = dto.Branch
	p.ImportURL = dto.ImportURL
}

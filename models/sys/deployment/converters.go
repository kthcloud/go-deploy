package deployment

import (
	"fmt"
	"go-deploy/models/dto/body"
	"log"
)

func (deployment *Deployment) ToDTO(url *string, storageManagerURL *string) body.DeploymentRead {
	var fullURL *string
	if url != nil {
		res := fmt.Sprintf("https://%s", *url)
		fullURL = &res
	}

	var fullStorageManagerURL *string
	if storageManagerURL != nil {
		res := fmt.Sprintf("https://%s", *storageManagerURL)
		fullStorageManagerURL = &res
	}

	app := deployment.GetMainApp()
	if app == nil {
		log.Println("main app not found in deployment", deployment.ID)
		app = &App{}
	}

	if app.Envs == nil {
		app.Envs = []Env{}
	}

	envs := make([]body.Env, len(app.Envs))
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

	return body.DeploymentRead{
		ID:      deployment.ID,
		Name:    deployment.Name,
		OwnerID: deployment.OwnerID,
		Zone:    deployment.Zone,

		URL:          fullURL,
		Envs:         envs,
		Volumes:      volumes,
		InitCommands: app.InitCommands,
		Private:      app.Private,

		Status:     deployment.StatusMessage,
		PingResult: pingResult,

		Integrations: integrations,

		StorageURL: fullStorageManagerURL,
	}
}

func (g *GitHubRepository) ToDTO() body.GitHubRepository {
	return body.GitHubRepository{
		ID:   g.ID,
		Name: g.Name,
	}
}

func (p *UpdateParams) FromDTO(dto *body.DeploymentUpdate) {
	if dto.Envs != nil {
		envs := make([]Env, len(*dto.Envs))
		for i, env := range *dto.Envs {
			envs[i] = Env{
				Name:  env.Name,
				Value: env.Value,
			}
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
}

func (p *CreateParams) FromDTO(dto *body.DeploymentCreate, fallbackZone *string, fallbackPort int) {
	p.Name = dto.Name

	if dto.InternalPort == nil {
		p.InternalPort = fallbackPort
	} else {
		p.InternalPort = *dto.InternalPort
	}
	p.Private = dto.Private
	p.Envs = make([]Env, len(dto.Envs))
	for i, env := range dto.Envs {
		p.Envs[i] = Env{
			Name:  env.Name,
			Value: env.Value,
		}
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

	if dto.GitHub != nil {
		p.GitHub = &GitHubCreateParams{
			Token:        dto.GitHub.Token,
			RepositoryID: dto.GitHub.RepositoryID,
		}
	}

	if dto.Zone != nil {
		p.Zone = *dto.Zone
	} else {
		p.Zone = *fallbackZone
	}
}

func (p *BuildParams) FromDTO(dto *body.DeploymentBuild) {
	p.Tag = dto.Tag
	p.Branch = dto.Branch
	p.ImportURL = dto.ImportURL
}

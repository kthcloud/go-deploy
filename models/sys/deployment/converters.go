package deployment

import (
	"fmt"
	"go-deploy/models/dto/body"
)

func (deployment *Deployment) ToDTO(url *string) body.DeploymentRead {
	var fullURL *string
	if url != nil {
		res := fmt.Sprintf("https://%s", *url)
		fullURL = &res
	}

	if deployment.Envs == nil {
		deployment.Envs = []Env{}
	}

	envs := make([]body.Env, len(deployment.Envs))
	for i, env := range deployment.Envs {
		envs[i] = body.Env{
			Name:  env.Name,
			Value: env.Value,
		}
	}

	volumes := make([]body.Volume, len(deployment.Volumes))
	for i, volume := range deployment.Volumes {
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

	if deployment.InitCommands == nil {
		deployment.InitCommands = make([]string, 0)
	}

	var pingResult *int
	if deployment.PingResult != 0 {
		pingResult = &deployment.PingResult
	}

	return body.DeploymentRead{
		ID:      deployment.ID,
		Name:    deployment.Name,
		OwnerID: deployment.OwnerID,
		Zone:    deployment.Zone,

		URL:          fullURL,
		Envs:         envs,
		Volumes:      volumes,
		InitCommands: deployment.InitCommands,
		Private:      deployment.Private,

		Status:     deployment.StatusMessage,
		PingResult: pingResult,

		Integrations: integrations,
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

	p.Private = dto.Private
	p.ExtraDomains = dto.ExtraDomains
}

func (p *CreateParams) FromDTO(dto *body.DeploymentCreate, fallbackZone *string) {
	p.Name = dto.Name
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

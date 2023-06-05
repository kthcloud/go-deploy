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

	var build *body.Build
	if deployment.Subsystems.GitLab.LastBuild.ID != 0 {
		lb := &deployment.Subsystems.GitLab.LastBuild
		build = &body.Build{
			Trace:     lb.Trace,
			Status:    lb.Status,
			Stage:     lb.Stage,
			CreatedAt: lb.CreatedAt,
		}
	}

	return body.DeploymentRead{
		ID:      deployment.ID,
		Name:    deployment.Name,
		OwnerID: deployment.OwnerID,
		Status:  deployment.StatusMessage,
		URL:     fullURL,
		Envs:    envs,
		Build:   build,
		Private: deployment.Private,
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

func (p *CreateParams) FromDTO(dto *body.DeploymentCreate) {
	p.Name = dto.Name
	p.Private = dto.Private
	p.Envs = make([]Env, len(dto.Envs))
	for i, env := range dto.Envs {
		p.Envs[i] = Env{
			Name:  env.Name,
			Value: env.Value,
		}
	}

	if dto.GitHub != nil {
		p.GitHub = &GitHubCreateParams{
			Token:        dto.GitHub.Token,
			RepositoryID: dto.GitHub.RepositoryID,
		}
	}
}

func (p *BuildParams) FromDTO(dto *body.DeploymentBuild) {
	p.Tag = dto.Tag
	p.Branch = dto.Branch
	p.ImportURL = dto.ImportURL
}

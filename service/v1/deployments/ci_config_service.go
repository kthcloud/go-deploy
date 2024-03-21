package deployments

import (
	"fmt"
	"go-deploy/dto/v1/body"
	"go-deploy/models/model"
	"go-deploy/pkg/config"
	"go-deploy/service/errors"
	"go-deploy/service/v1/deployments/opts"
	"go-deploy/utils/subsystemutils"
	"gopkg.in/yaml.v3"
)

// GetCiConfig returns the CI config for the deployment.
//
// It returns an error if the deployment is not found, or if the deployment is not ready.
// It returns nil if the deployment is not a custom deployment.
func (c *Client) GetCiConfig(id string) (*body.CiConfig, error) {
	deployment, err := c.Get(id, opts.GetOpts{Shared: true})
	if err != nil {
		return nil, err
	}

	if deployment == nil {
		return nil, errors.DeploymentNotFoundErr
	}

	if !deployment.Ready() {
		return nil, nil
	}

	if deployment.Type != model.DeploymentTypeCustom {
		return nil, nil
	}

	tag := fmt.Sprintf("%s/%s/%s",
		config.Config.Registry.URL,
		subsystemutils.GetPrefixedName(deployment.OwnerID),
		deployment.Name,
	)

	username := deployment.Subsystems.Harbor.Robot.HarborName
	password := deployment.Subsystems.Harbor.Robot.Secret

	githubCiConfig := model.GithubActionConfig{
		Name: "kthcloud-ci",
		On:   model.On{Push: model.Push{Branches: []string{"main"}}},
		Jobs: model.Jobs{Docker: model.Docker{
			RunsOn: "ubuntu-latest",
			Steps: []model.Steps{
				{
					Name: "Login to Docker Hub",
					Uses: "docker/login-action@v3",
					With: model.With{
						Registry: config.Config.Registry.URL,
						Username: username,
						Password: password,
					},
				},
				{
					Name: "Build and push",
					Uses: "docker/build-push-action@v5",
					With: model.With{
						Push: true,
						Tags: tag,
					},
				},
			},
		}},
	}

	marshalledConfig, err := yaml.Marshal(githubCiConfig)
	if err != nil {
		return nil, err
	}

	ciConfig := body.CiConfig{Config: string(marshalledConfig)}
	return &ciConfig, nil
}

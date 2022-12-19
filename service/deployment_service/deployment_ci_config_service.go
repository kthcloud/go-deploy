package deployment_service

import (
	"fmt"
	"go-deploy/models"
	"go-deploy/models/dto"
	"go-deploy/pkg/conf"
	"go-deploy/utils/subsystemutils"
	"gopkg.in/yaml.v2"
)

func GetCIConfig(userId, deploymentID string) (*dto.CIConfig, error) {
	deployment, err := Get(userId, deploymentID)
	if err != nil {
		return nil, err
	}

	if deployment == nil {
		return nil, nil
	}

	if !deployment.Ready() {
		return nil, nil
	}

	tag := fmt.Sprintf("%s/%s/%s",
		conf.Env.DockerRegistry.Url,
		subsystemutils.GetPrefixedName(deployment.Name),
		deployment.Name,
	)

	username := deployment.Subsystems.Harbor.Robot.Name
	password := deployment.Subsystems.Harbor.Robot.Secret

	config := models.GithubActionConfig{
		Name: "kthcloud-ci",
		On:   models.On{Push: models.Push{Branches: []string{"main"}}},
		Jobs: models.Jobs{Docker: models.Docker{
			RunsOn: "ubuntu-latest",
			Steps: []models.Steps{
				{
					Name: "Set up QEMU",
					Uses: "docker/setup-qemu-action@v2",
				},
				{
					Name: "Set up Docker Buildx",
					Uses: "docker/setup-buildx-action@v2",
				},
				{
					Name: "Login to Docker Hub",
					Uses: "docker/login-action@v2",
					With: models.With{
						Registry: conf.Env.DockerRegistry.Url,
						Username: username,
						Password: password,
					},
				},
				{
					Name: "Build and push",
					Uses: "docker/build-push-action@v3",
					With: models.With{
						Push: true,
						Tags: tag,
					},
				},
			},
		}},
	}

	marshalledConfig, err := yaml.Marshal(config)
	if err != nil {
		return nil, err
	}

	ciConfig := dto.CIConfig{Config: string(marshalledConfig)}
	return &ciConfig, nil
}

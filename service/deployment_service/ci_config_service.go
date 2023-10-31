package deployment_service

import (
	"fmt"
	"go-deploy/models/dto/body"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/pkg/config"
	"go-deploy/service"
	"go-deploy/utils/subsystemutils"
	"gopkg.in/yaml.v2"
	"log"
)

func GetCIConfig(deploymentID string, auth *service.AuthInfo) (*body.CiConfig, error) {
	deployment, err := GetByIdAuth(deploymentID, auth)
	if err != nil {
		return nil, err
	}

	if deployment == nil {
		log.Println("deployment", deploymentID, "not found for ci config fetch. assuming it was deleted")
		return nil, nil
	}

	if !deployment.Ready() {
		return nil, nil
	}

	if deployment.Type != deploymentModel.TypeCustom {
		return nil, nil
	}

	tag := fmt.Sprintf("%s/%s/%s",
		config.Config.Registry.URL,
		subsystemutils.GetPrefixedName(auth.UserID),
		deployment.Name,
	)

	username := deployment.Subsystems.Harbor.Robot.HarborName
	password := deployment.Subsystems.Harbor.Robot.Secret

	config := deploymentModel.GithubActionConfig{
		Name: "kthcloud-ci",
		On:   deploymentModel.On{Push: deploymentModel.Push{Branches: []string{"main"}}},
		Jobs: deploymentModel.Jobs{Docker: deploymentModel.Docker{
			RunsOn: "ubuntu-latest",
			Steps: []deploymentModel.Steps{
				{
					Name: "Login to Docker Hub",
					Uses: "docker/login-action@v2",
					With: deploymentModel.With{
						Registry: config.Config.Registry.URL,
						Username: username,
						Password: password,
					},
				},
				{
					Name: "Build and push",
					Uses: "docker/build-push-action@v3",
					With: deploymentModel.With{
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

	ciConfig := body.CiConfig{Config: string(marshalledConfig)}
	return &ciConfig, nil
}

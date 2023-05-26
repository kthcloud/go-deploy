package internal_service

import (
	"fmt"
	"github.com/google/uuid"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/gitlab"
	"go-deploy/pkg/subsystems/gitlab/models"
	"go-deploy/utils/subsystemutils"
	"log"
)

func CreateBuild(id string, params *deploymentModel.BuildParams) error {
	log.Println("creating build with gitlab")

	makeError := func(err error) error {
		return fmt.Errorf("failed to build deployment with gitlab. details: %s", err)
	}

	deployment, err := deploymentModel.GetByID(id)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", id, "not found for gitlab build. assuming it was deleted")
		return nil
	}

	client, err := gitlab.New(&gitlab.ClientConf{
		URL:   conf.Env.GitLab.URL,
		Token: conf.Env.GitLab.Token,
	})

	if err != nil {
		return makeError(err)
	}

	public := &models.ProjectPublic{
		Name:      deployment.Name + "-" + uuid.NewString(),
		ImportURL: params.ImportURL,
	}

	projectID, err := client.CreateProject(public)
	if err != nil {
		return makeError(err)
	}

	err = client.AttachCiFile(projectID,
		params.Branch,
		deploymentModel.GitLabCiConfig{
			Build: deploymentModel.Build{
				Image: "docker:20.10.16",
				Stage: "build",
				Services: []string{
					"docker:20.10.16-dind",
				},
				Variables: map[string]string{
					"CI_REGISTRY":          conf.Env.DockerRegistry.URL,
					"CI_REGISTRY_IMAGE":    conf.Env.DockerRegistry.URL + "/" + deployment.OwnerID + "/" + subsystemutils.GetPrefixedName(deployment.Name),
					"CI_COMMIT_REF_SLUG":   params.Tag,
					"CI_REGISTRY_USER":     deployment.Subsystems.Harbor.Robot.Name,
					"CI_REGISTRY_PASSWORD": deployment.Subsystems.Harbor.Robot.Secret,
				},
				Script: []string{
					"docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY",
					"docker build --pull -t $CI_REGISTRY_IMAGE:$CI_COMMIT_REF_SLUG .",
					"docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_REF_SLUG",
				},
			},
		},
	)

	if err != nil {
		return makeError(err)
	}

	return nil
}

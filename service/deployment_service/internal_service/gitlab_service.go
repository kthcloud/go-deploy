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
	"strings"
	"time"
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

	defer func() {
		err = client.DeleteProject(projectID)
		if err != nil {
			log.Println("failed to delete gitlab project", projectID, "after build. details:", err)
		}
	}()

	escapedHarborName := strings.Replace(deployment.Subsystems.Harbor.Robot.HarborName, "$", "$$", -1)

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
					"CI_REGISTRY_IMAGE":    conf.Env.DockerRegistry.URL + "/" + subsystemutils.GetPrefixedName(deployment.OwnerID) + "/" + deployment.Name,
					"CI_COMMIT_REF_SLUG":   params.Tag,
					"CI_REGISTRY_USER":     escapedHarborName,
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

	var lastJob *models.JobPublic
	for {
		lastJob, err = client.ReadLastJob(projectID)
		if err != nil {
			return makeError(err)
		}

		if lastJob != nil {
			break
		}
	}

	for {
		var trace string
		trace, err = client.GetJobTrace(projectID, lastJob.ID)
		if err != nil {
			return makeError(err)
		}

		traceSlice := strings.Split(trace, "\n")

		err = updateGitLabBuild(id, lastJob, traceSlice)
		if err != nil {
			return makeError(err)
		}

		lastJob, err = client.ReadLastJob(projectID)
		if err != nil {
			return makeError(err)
		}
		if lastJob == nil {
			log.Println("build job disappeared in gitlab. assuming it was deleted")
			break
		}

		if lastJob.Status == "success" || lastJob.Status == "failed" {
			err = updateGitLabBuild(id, lastJob, traceSlice)
			if err != nil {
				return makeError(err)
			}
			break
		}

		time.Sleep(1 * time.Second)
	}

	log.Println("build finished with gitlab")

	return nil
}

func updateGitLabBuild(deploymentID string, lastJob *models.JobPublic, trace []string) error {
	return deploymentModel.UpdateGitLabBuild(deploymentID, deploymentModel.GitLabBuild{
		ID:        lastJob.ID,
		ProjectID: lastJob.ProjectID,
		Trace:     trace,
		Status:    lastJob.Status,
		Stage:     lastJob.Stage,
		CreatedAt: lastJob.CreatedAt,
	})
}

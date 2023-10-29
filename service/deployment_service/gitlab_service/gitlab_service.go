package gitlab_service

import (
	"fmt"
	"github.com/google/uuid"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/gitlab"
	"go-deploy/pkg/subsystems/gitlab/models"
	"go-deploy/utils"
	"go-deploy/utils/subsystemutils"
	"log"
	"strings"
	"time"
)

func CreateBuild(ids []string, params *deploymentModel.BuildParams) error {
	log.Println("creating build with gitlab for", len(ids), "deployments")

	makeError := func(err error) error {
		return fmt.Errorf("failed to build deployment with gitlab. details: %w", err)
	}

	client, err := gitlab.New(&gitlab.ClientConf{
		URL:   conf.Env.GitLab.URL,
		Token: conf.Env.GitLab.Token,
	})

	name := params.Name
	if name == "" {
		name = "build"
	}

	public := &models.ProjectPublic{
		Name:      fmt.Sprintf("%s-%s", name, uuid.New().String()),
		ImportURL: params.ImportURL,
	}

	projectID, err := client.CreateProject(public)
	if err != nil {
		return makeError(err)
	}

	defer func() {
		err = client.DeleteProject(projectID)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to delete gitlab project %d after build. details: %w", projectID, err))
		}
	}()

	script := []string{
		"docker build --pull -t build .",
	}

	for _, id := range ids {
		deployment, err := deploymentModel.New().GetByID(id)
		if err != nil {
			return makeError(err)
		}

		// harbor name contains $ which is a special character in gitlab ci, so we need to escape it with $$.
		user := strings.Replace(deployment.Subsystems.Harbor.Robot.HarborName, "$", "\\$", -1)
		password := deployment.Subsystems.Harbor.Robot.Secret
		registry := conf.Env.Registry.URL
		path := fmt.Sprintf("%s/%s/%s", registry, subsystemutils.GetPrefixedName(deployment.OwnerID), deployment.Name)

		script = append(script, fmt.Sprintf("docker login -u %s -p %s %s", user, password, registry))
		script = append(script, fmt.Sprintf("docker tag %s %s", "build", path))
		script = append(script, fmt.Sprintf("docker push %s", path))
	}

	err = client.AttachCiFile(projectID,
		params.Branch,
		deploymentModel.GitLabCiConfig{
			Build: deploymentModel.Build{
				Image: "docker:24.0.5",
				Stage: "build",
				Services: []string{
					"docker:24.0.5-dind",
				},
				BeforeScript: []string{
					"docker info",
				},
				Script: script,
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

		for _, id := range ids {
			err = updateGitLabBuild(id, lastJob, traceSlice)
			if err != nil {
				return makeError(err)
			}
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
			for _, id := range ids {
				err = updateGitLabBuild(id, lastJob, traceSlice)
				if err != nil {
					return makeError(err)
				}
			}
			break
		}

		time.Sleep(1 * time.Second)
	}

	log.Println("build finished with gitlab for", len(ids), "deployments")
	return nil
}

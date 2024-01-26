package gitlab_service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/gitlab"
	"go-deploy/pkg/subsystems/gitlab/models"
	"go-deploy/utils"
	"go-deploy/utils/subsystemutils"
	"log"
	"strings"
	"time"
)

const (
	// BuildStatusPending is the status of a pending build.
	BuildStatusPending = "pending"
	// BuildStatusRunning is the status of a running build.
	BuildStatusRunning = "running"

	// JobStatusSuccess is the status of a successful job.
	JobStatusSuccess = "success"
	// JobStatusFailed is the status of a failed job.
	JobStatusFailed = "failed"
)

// CreateBuild creates a build for the given deployment.
//
// It creates a temporary GitLab project, imports the git repository to it, and creates a build job.
// Once the build job is finished, it will push to the Container registry; the temporary GitLab project is deleted
//
// Since this is connected to Harbor, the push will trigger a Harbor webhook, and the deployment will be restarted.
func CreateBuild(ids []string, params *deploymentModels.BuildParams) error {
	log.Println("creating build with gitlab for", len(ids), "deployments")

	makeError := func(err error) error {
		return fmt.Errorf("failed to build deployment with gitlab. details: %w", err)
	}

	client, err := gitlab.New(&gitlab.ClientConf{
		URL:   config.Config.GitLab.URL,
		Token: config.Config.GitLab.Token,
	})

	name := params.Name
	if name == "" {
		name = "build"
	}

	public := &models.ProjectPublic{
		Name:      fmt.Sprintf("%s-%s", name, uuid.New().String()),
		ImportURL: params.ImportURL,
	}

	project, err := client.CreateProject(public)
	if err != nil {
		return makeError(err)
	}

	defer func() {
		err = client.DeleteProject(project.ID)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to delete gitlab project %d after build. details: %w", project.ID, err))
		}
	}()

	script := []string{
		"docker build --pull -t build .",
	}

	for _, id := range ids {
		deployment, err := deploymentModels.New().GetByID(id)
		if err != nil {
			return makeError(err)
		}

		// harbor name contains $ which is a special character in gitlab ci, so we need to escape it with $$.
		user := strings.Replace(deployment.Subsystems.Harbor.Robot.HarborName, "$", "\\$", -1)
		password := deployment.Subsystems.Harbor.Robot.Secret
		registry := config.Config.Registry.URL
		path := fmt.Sprintf("%s/%s/%s", registry, subsystemutils.GetPrefixedName(deployment.OwnerID), deployment.Name)

		script = append(script, fmt.Sprintf("docker login -u %s -p %s %s", user, password, registry))
		script = append(script, fmt.Sprintf("docker tag %s %s", "build", path))
		script = append(script, fmt.Sprintf("docker push %s", path))
	}

	err = client.AttachCiFile(project.ID,
		params.Branch,
		deploymentModels.GitLabCiConfig{
			Build: deploymentModels.Build{
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
		lastJob, err = client.ReadLastJob(project.ID)
		if err != nil {
			return makeError(err)
		}

		if lastJob != nil {
			break
		}
	}

	for {
		var trace string
		trace, err = client.GetJobTrace(project.ID, lastJob.ID)
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

		lastJob, err = client.ReadLastJob(project.ID)
		if err != nil {
			return makeError(err)
		}
		if lastJob == nil {
			log.Println("build job disappeared in gitlab. assuming it was deleted")
			break
		}

		if lastJob.Status == JobStatusSuccess || lastJob.Status == JobStatusFailed {
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

// SetupLogStream sets up a continuous log stream for the given deployment.
//
// It will continuously check the GitLab build status and read the build trace.
// If the build is running, it will read the trace from the last read line.
// If the build is finished, it will read the trace from the beginning.
//
// It will call the handler function for each line read.
// The handler function is expected to handle the line and the time it was read.
func SetupLogStream(ctx context.Context, deploymentID string, handler func(string, time.Time)) error {
	buildID := 0
	readLines := 0

	go func() {
		for {
			time.Sleep(300 * time.Millisecond)

			select {
			case <-ctx.Done():
				return
			default:
				build, err := deploymentModels.New().GetLastGitLabBuild(deploymentID)
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to get last gitlab build when setting up continuous log stream. details: %w", err))
					return
				}

				if build == nil {
					continue
				}

				if build.ID == 0 {
					continue
				}

				if buildID != build.ID {
					buildID = build.ID
					readLines = 0
				}

				if build.Status == BuildStatusRunning || build.Status == BuildStatusPending {
					for _, line := range build.Trace[readLines:] {
						if line != "" {
							handler(line, time.Now())
						}
						readLines++
					}
				}
			}
		}
	}()
	return nil
}

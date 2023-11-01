package deployment_service

import (
	"context"
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/service"
	"go-deploy/utils"
	"go-deploy/utils/subsystemutils"
	"log"
	"time"
)

func SetupLogStream(deploymentID string, handler func(string), auth *service.AuthInfo) (context.Context, error) {
	deployment, err := GetByIdAuth(deploymentID, auth)
	if err != nil {
		return nil, err
	}

	if deployment == nil {
		return nil, DeploymentNotFoundErr
	}

	if deployment.BeingDeleted() {
		log.Println("deployment", deploymentID, "is being deleted. not setting up log stream")
		return nil, nil
	}

	zone := config.Config.Deployment.GetZone(deployment.Zone)
	if zone == nil {
		return nil, fmt.Errorf("zone %s not found", deployment.Zone)
	}

	ctx := context.Background()

	k8sClient, err := k8s.New(zone.Client, subsystemutils.GetPrefixedName(deployment.OwnerID))
	if err != nil {
		return nil, err
	}

	k8sDeployment := deployment.Subsystems.K8s.GetDeployment(deployment.Name)
	if service.NotCreated(k8sDeployment) {
		log.Println("deployment", deploymentID, "not found in k8s when setting up log stream. assuming it was deleted")
		return nil, DeploymentNotFoundErr
	}

	ssK8s := deployment.Subsystems.K8s
	err = k8sClient.SetupLogStream(ctx, ssK8s.Namespace.Name, k8sDeployment.ID, handler)
	if err != nil {
		ctx.Done()
		return nil, err
	}

	err = setupContinuousGitLabLogStream(ctx, deploymentID, handler)
	if err != nil {
		return nil, err
	}

	return ctx, nil
}

func setupContinuousGitLabLogStream(ctx context.Context, deploymentID string, handler func(string)) error {
	buildID := 0
	readRows := 0

	go func() {
		for {
			time.Sleep(300 * time.Millisecond)

			select {
			case <-ctx.Done():
				return
			default:
				build, err := deploymentModel.New().GetLastGitLabBuild(deploymentID)
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
					readRows = 0
				}

				if build.Status == "pending" || build.Status == "running" {
					for _, row := range build.Trace[readRows:] {
						if row != "" {
							handler(row)
						}
						readRows++
					}
				}
			}
		}
	}()
	return nil
}

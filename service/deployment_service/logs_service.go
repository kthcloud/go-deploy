package deployment_service

import (
	"context"
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/models/sys/key_value"
	"go-deploy/pkg/config"
	"go-deploy/pkg/metrics"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/service"
	"go-deploy/utils"
	"go-deploy/utils/subsystemutils"
	"log"
	"time"
)

const (
	MessageSourcePod     = "pod"
	MessageSourceBuild   = "build"
	MessageSourceControl = "control"
)

func SetupLogStream(ctx context.Context, deploymentID string, handler func(string, string, string), auth *service.AuthInfo) error {
	deployment, err := GetByIdAuth(deploymentID, auth)
	if err != nil {
		return err
	}

	if deployment == nil {
		return DeploymentNotFoundErr
	}

	if deployment.BeingDeleted() {
		log.Println("deployment", deploymentID, "is being deleted. not setting up log stream")
		return nil
	}

	zone := config.Config.Deployment.GetZone(deployment.Zone)
	if zone == nil {
		return fmt.Errorf("zone %s not found", deployment.Zone)
	}

	k8sClient, err := k8s.New(zone.Client, subsystemutils.GetPrefixedName(deployment.OwnerID))
	if err != nil {
		return err
	}

	k8sDeployment := deployment.Subsystems.K8s.GetDeployment(deployment.Name)
	if service.NotCreated(k8sDeployment) {
		log.Println("deployment", deploymentID, "not found in k8s when setting up log stream. assuming it was deleted")
		return DeploymentNotFoundErr
	}

	ssK8s := deployment.Subsystems.K8s
	err = k8sClient.SetupLogStream(ctx, ssK8s.Namespace.Name, k8sDeployment.ID, func(podNumber int, line string) {
		if line == k8s.ControlMessage {
			handler(MessageSourceControl, "[control]", line)
			return
		}
		handler(MessageSourcePod, fmt.Sprintf("[pod %d]", podNumber), line)
	})
	if err != nil {
		return err
	}

	go func() {
		err = key_value.New().Incr(metrics.KeyThreadsLog)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to increment log thread when setting up continuous log stream. details: %w", err))
		}

		for {
			time.Sleep(300 * time.Millisecond)
			if ctx.Err() != nil {
				err = key_value.New().Decr(metrics.KeyThreadsLog)
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to decrement log thread when setting up continuous log stream. details: %w", err))
				}
				return
			}
		}
	}()

	err = setupContinuousGitLabLogStream(ctx, deploymentID, handler)
	if err != nil {
		return err
	}

	return nil
}

func setupContinuousGitLabLogStream(ctx context.Context, deploymentID string, handler func(string, string, string)) error {
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
							handler(MessageSourceBuild, "[build]", row)
						}
						readRows++
					}
				}
			}
		}
	}()
	return nil
}

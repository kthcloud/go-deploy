package confirm

import (
	"context"
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	jobModel "go-deploy/models/sys/job"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/utils"
	"golang.org/x/exp/slices"
	"log"
	"time"
)

func deploymentConfirmer(ctx context.Context) {
	defer log.Println("deploymentConfirmer stopped")

	for {
		select {
		case <-time.After(5 * time.Second):
			beingCreated, _ := deploymentModel.New().ListByActivity(deploymentModel.ActivityBeingCreated)
			for _, deployment := range beingCreated {
				created := DeploymentCreated(&deployment)
				if created {
					log.Printf("marking deployment %s as created\n", deployment.ID)
					_ = deploymentModel.New().RemoveActivity(deployment.ID, deploymentModel.ActivityBeingCreated)
				}
			}

			beingDeleted, _ := deploymentModel.New().ListByActivity(deploymentModel.ActivityBeingDeleted)
			for _, deployment := range beingDeleted {
				deleted := DeploymentDeleted(&deployment)
				if !deleted {
					continue
				}

				relatedJobs, err := jobModel.New().ExcludeScheduled().GetByArgs(map[string]interface{}{
					"id": deployment.ID,
				})

				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to get related jobs when confirming deployment deleting. details: %w", err))
					continue
				}

				allFinished := slices.IndexFunc(relatedJobs, func(j jobModel.Job) bool {
					return j.Status != jobModel.StatusCompleted &&
						j.Status != jobModel.StatusTerminated
				}) == -1

				if allFinished {
					log.Printf("marking deployment %s as deleted\n", deployment.ID)
					_ = deploymentModel.New().DeleteByID(deployment.ID)
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

func vmConfirmer(ctx context.Context) {
	defer log.Println("vmConfirmer stopped")

	for {
		select {
		case <-time.After(5 * time.Second):
			beingCreated, err := vmModel.New().ListByActivity(vmModel.ActivityBeingCreated)
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to get vms being created. details: %w", err))
			}

			for _, vm := range beingCreated {
				created := VmCreated(&vm)
				if created {
					log.Printf("marking vm %s as created\n", vm.ID)
					_ = vmModel.New().RemoveActivity(vm.ID, vmModel.ActivityBeingCreated)
				}
			}

			beingDeleted, err := vmModel.New().ListByActivity(vmModel.ActivityBeingDeleted)
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to get vms being deleted. details: %w", err))
			}

			for _, vm := range beingDeleted {
				deleted := VmDeleted(&vm)
				if !deleted {
					continue
				}

				relatedJobs, err := jobModel.New().ExcludeScheduled().GetByArgs(map[string]interface{}{
					"id": vm.ID,
				})

				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to get related jobs when confirming vm deleting. details: %w", err))
					continue
				}

				allFinished := slices.IndexFunc(relatedJobs, func(j jobModel.Job) bool {
					return j.Status != jobModel.StatusCompleted &&
						j.Status != jobModel.StatusTerminated
				}) == -1

				if allFinished {
					log.Printf("marking vm %s as deleted\n", vm.ID)
					_ = vmModel.New().DeleteByID(vm.ID)
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

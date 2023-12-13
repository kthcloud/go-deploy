package confirm

import (
	"context"
	"fmt"
	deploymentModels "go-deploy/models/sys/deployment"
	jobModels "go-deploy/models/sys/job"
	smModels "go-deploy/models/sys/sm"
	vmModels "go-deploy/models/sys/vm"
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
			beingCreated, _ := deploymentModels.New().WithActivities(deploymentModels.ActivityBeingCreated).List()
			for _, deployment := range beingCreated {
				created := DeploymentCreated(&deployment)
				if created {
					log.Printf("marking deployment %s as created\n", deployment.ID)
					_ = deploymentModels.New().RemoveActivity(deployment.ID, deploymentModels.ActivityBeingCreated)
				}
			}

			beingDeleted, _ := deploymentModels.New().WithActivities(deploymentModels.ActivityBeingDeleted).List()
			for _, deployment := range beingDeleted {
				deleted := DeploymentDeleted(&deployment)
				if !deleted {
					continue
				}

				relatedJobs, err := jobModels.New().ExcludeScheduled().GetByArgs(map[string]interface{}{
					"id": deployment.ID,
				})

				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to get related jobs when confirming deployment deleting. details: %w", err))
					continue
				}

				allFinished := slices.IndexFunc(relatedJobs, func(j jobModels.Job) bool {
					return j.Status != jobModels.StatusCompleted &&
						j.Status != jobModels.StatusTerminated
				}) == -1

				if allFinished {
					log.Printf("marking deployment %s as deleted\n", deployment.ID)
					_ = deploymentModels.New().DeleteByID(deployment.ID)
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

func smConfirmer(ctx context.Context) {
	defer log.Println("smConfirmer stopped")

	for {
		select {
		case <-time.After(5 * time.Second):
			beingCreated, err := smModels.New().WithActivities(smModels.ActivityBeingCreated).List()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to get sms being created. details: %w", err))
			}

			for _, sm := range beingCreated {
				created := SmCreated(&sm)
				if created {
					log.Printf("marking sm %s as created\n", sm.ID)
					_ = smModels.New().RemoveActivity(sm.ID, smModels.ActivityBeingCreated)
				}
			}

			beingDeleted, err := smModels.New().WithActivities(smModels.ActivityBeingDeleted).List()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to get sms being deleted. details: %w", err))
			}

			for _, sm := range beingDeleted {
				deleted := SmDeleted(&sm)
				if !deleted {
					continue
				}

				relatedJobs, err := jobModels.New().ExcludeScheduled().GetByArgs(map[string]interface{}{
					"id": sm.ID,
				})

				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to get related jobs when confirming sm deleting. details: %w", err))
					continue
				}

				allFinished := slices.IndexFunc(relatedJobs, func(j jobModels.Job) bool {
					return j.Status != jobModels.StatusCompleted &&
						j.Status != jobModels.StatusTerminated
				}) == -1

				if allFinished {
					log.Printf("marking sm %s as deleted\n", sm.ID)
					_ = smModels.New().DeleteByID(sm.ID)
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
			beingCreated, err := vmModels.New().WithActivities(vmModels.ActivityBeingCreated).List()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to get vms being created. details: %w", err))
			}

			for _, vm := range beingCreated {
				created := VmCreated(&vm)
				if created {
					log.Printf("marking vm %s as created\n", vm.ID)
					_ = vmModels.New().RemoveActivity(vm.ID, vmModels.ActivityBeingCreated)
				}
			}

			beingDeleted, err := vmModels.New().WithActivities(vmModels.ActivityBeingDeleted).List()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to get vms being deleted. details: %w", err))
			}

			for _, vm := range beingDeleted {
				deleted := VmDeleted(&vm)
				if !deleted {
					continue
				}

				relatedJobs, err := jobModels.New().ExcludeScheduled().GetByArgs(map[string]interface{}{
					"id": vm.ID,
				})

				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to get related jobs when confirming vm deleting. details: %w", err))
					continue
				}

				allFinished := slices.IndexFunc(relatedJobs, func(j jobModels.Job) bool {
					return j.Status != jobModels.StatusCompleted &&
						j.Status != jobModels.StatusTerminated
				}) == -1

				if allFinished {
					log.Printf("marking vm %s as deleted\n", vm.ID)
					_ = vmModels.New().DeleteByID(vm.ID)
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

package confirm

import (
	"context"
	"fmt"
	deploymentModels "go-deploy/models/sys/deployment"
	jobModels "go-deploy/models/sys/job"
	smModels "go-deploy/models/sys/sm"
	vmModels "go-deploy/models/sys/vm"
	"go-deploy/pkg/config"
	"go-deploy/pkg/workers"
	"go-deploy/utils"
	"golang.org/x/exp/slices"
	"log"
	"time"
)

// deploymentDeletionConfirmer is a worker that confirms deployment deletion.
func deploymentDeletionConfirmer(ctx context.Context) {
	defer workers.OnStop("deploymentDeletionConfirmer")

	reportTick := time.Tick(1 * time.Second)
	tick := time.Tick(3 * time.Second)

	for {
		select {
		case <-reportTick:
			workers.ReportUp("deploymentDeletionConfirmer")

		case <-tick:
			beingDeleted, _ := deploymentModels.New().WithActivities(deploymentModels.ActivityBeingDeleted).List()
			for _, deployment := range beingDeleted {
				deleted := DeploymentDeleted(&deployment)
				if !deleted {
					continue
				}

				relatedJobs, err := jobModels.New().ExcludeScheduled().FilterArgs("id", deployment.ID).List()
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

// smDeletionConfirmer is a worker that confirms SM deletion.
func smDeletionConfirmer(ctx context.Context) {
	defer workers.OnStop("smDeletionConfirmer")

	reportTick := time.Tick(1 * time.Second)
	tick := time.Tick(3 * time.Second)

	for {
		select {
		case <-reportTick:
			workers.ReportUp("smDeletionConfirmer")

		case <-tick:
			beingDeleted, err := smModels.New().WithActivities(smModels.ActivityBeingDeleted).List()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to get sms being deleted. details: %w", err))
			}

			for _, sm := range beingDeleted {
				deleted := SmDeleted(&sm)
				if !deleted {
					continue
				}

				relatedJobs, err := jobModels.New().ExcludeScheduled().FilterArgs("id", sm.ID).List()
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

// vmDeletionConfirmer is a worker that confirms VM deletion.
func vmDeletionConfirmer(ctx context.Context) {
	defer workers.OnStop("vmDeletionConfirmer")

	reportTick := time.Tick(1 * time.Second)
	tick := time.Tick(3 * time.Second)

	for {
		select {
		case <-reportTick:
			workers.ReportUp("vmDeletionConfirmer")

		case <-tick:
			beingDeleted, err := vmModels.New().WithActivities(vmModels.ActivityBeingDeleted).List()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to get vms being deleted. details: %w", err))
			}

			for _, vm := range beingDeleted {
				deleted := VmDeleted(&vm)
				if !deleted {
					continue
				}

				relatedJobs, err := jobModels.New().ExcludeScheduled().FilterArgs("id", vm.ID).List()
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

// customDomainConfirmer is a worker that confirms custom domain setup.
func customDomainConfirmer(ctx context.Context) {
	defer workers.OnStop("customDomainConfirmer")

	reportTick := time.Tick(1 * time.Second)
	tick := time.Tick(3 * time.Second)
	subDomain := config.Config.Deployment.CustomDomainTxtRecordSubdomain

	for {
		select {
		case <-reportTick:
			workers.ReportUp("customDomainConfirmer")

		case <-tick:
			deploymentsWithPendingCustomDomain, err := deploymentModels.New().WithPendingCustomDomain().List()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to get deployments with pending custom domain. details: %w", err))
				continue
			}

			for _, deployment := range deploymentsWithPendingCustomDomain {
				// Check if user has updated the DNS record with the custom domain secret
				// If yes, mark the deployment as custom domain confirmed
				cd := deployment.GetMainApp().CustomDomain
				if cd == nil {
					continue
				}

				exists, match, txtRecord, err := checkCustomDomain(cd.Domain, cd.Secret)
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to lookup TXT record under %s for custom domain %s for deployment %s. details: %w", subDomain, cd.Domain, deployment.ID, err))
					continue
				}

				if !exists {
					log.Printf("no TXT record found under %s when confirming custom domain %s for deployment %s\n", subDomain, cd.Domain, deployment.ID)
					continue
				}

				if !match {
					received := txtRecord
					expected := cd.Secret
					if len(received) > len(expected) {
						received = received[:len(expected)] + "..."
					}

					log.Printf("TXT record found under %s but secret does not match when confirming custom domain %s for deployment %s (received: %s, expected: %s)\n", subDomain, cd.Domain, deployment.ID, received, expected)
					err = deploymentModels.New().UpdateCustomDomainStatus(deployment.ID, deploymentModels.CustomDomainStatusVerificationFailed)
					if err != nil {
						utils.PrettyPrintError(fmt.Errorf("custom domain verification failed for deployment %s. details: %w", deployment.ID, err))
						continue
					}

					continue
				}

				log.Printf("marking custom domain %s as confirmed for deployment %s\n", cd.Domain, deployment.ID)
				err = deploymentModels.New().UpdateCustomDomainStatus(deployment.ID, deploymentModels.CustomDomainStatusActive)
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to mark custom domain %s as confirmed for deployment %s. details: %w", cd.Domain, deployment.ID, err))
					continue
				}
			}

			vmsWithPendingCustomDomain, err := vmModels.New().ListWithAnyPendingCustomDomain()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to get vms with pending custom domain. details: %w", err))
				continue
			}

			for _, vm := range vmsWithPendingCustomDomain {
				// Check if user has updated the DNS record with the custom domain secret
				// If yes, mark the deployment as custom domain confirmed
				for portName, port := range vm.PortMap {
					if port.HttpProxy == nil || port.HttpProxy.CustomDomain == nil {
						continue
					}

					cd := port.HttpProxy.CustomDomain
					if cd == nil {
						continue
					}

					exists, match, txtRecord, err := checkCustomDomain(cd.Domain, cd.Secret)
					if err != nil {
						utils.PrettyPrintError(fmt.Errorf("failed to check custom domain %s for vm %s. details: %w", cd.Domain, vm.ID, err))
						continue
					}

					if !exists {
						log.Printf("no TXT record found under %s when confirming custom domain %s for vm %s\n", subDomain, cd.Domain, vm.ID)
						continue
					}

					if !match {
						received := txtRecord
						expected := cd.Secret
						if len(received) > len(expected) {
							received = received[:len(expected)] + "..."
						}

						log.Printf("TXT record found under %s but secret does not match when confirming custom domain %s for vm %s (received: %s, expected: %s)\n", subDomain, cd.Domain, vm.ID, received, expected)
						err = vmModels.New().UpdateCustomDomainStatus(vm.ID, portName, vmModels.CustomDomainStatusVerificationFailed)
						if err != nil {
							utils.PrettyPrintError(fmt.Errorf("custom domain verification failed for vm %s. details: %w", vm.ID, err))
							continue
						}

						continue
					}

					log.Printf("marking custom domain %s as confirmed for vm %s\n", cd.Domain, vm.ID)
					err = vmModels.New().UpdateCustomDomainStatus(vm.ID, portName, vmModels.CustomDomainStatusActive)
					if err != nil {
						utils.PrettyPrintError(fmt.Errorf("failed to mark custom domain %s as confirmed for vm %s. details: %w", cd.Domain, vm.ID, err))
						continue
					}
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

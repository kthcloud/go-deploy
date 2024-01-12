package confirm

import (
	"context"
	"errors"
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
	"net"
	"time"
)

func deploymentConfirmer(ctx context.Context) {
	defer log.Println("deploymentConfirmer stopped")
	for {
		select {
		case <-time.After(1 * time.Second):
			workers.ReportStatus("deploymentConfirmer")

		case <-time.After(5 * time.Second):
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

func customDomainConfirmer(ctx context.Context) {
	defer log.Println("customDomainConfirmer stopped")

	for {
		select {
		case <-time.After(1 * time.Second):
			workers.ReportStatus("customDomainConfirmer")

		case <-time.After(1 * time.Second):
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

				match, err := checkCustomDomain(cd.Domain, cd.Secret)
				if err != nil {
					log.Printf("failed to check custom domain %s for deployment %s. details: %v\n", cd.Domain, deployment.ID, err)
				}

				if !match {
					err = deploymentModels.New().UpdateCustomDomainStatus(deployment.ID, deploymentModels.CustomDomainStatusVerificationFailed)
					if err != nil {
						utils.PrettyPrintError(fmt.Errorf("custom domain verification failed. details: %w", err))
						continue
					}
				}

				log.Printf("marking custom domain %s as confirmed for deployment %s\n", cd.Domain, deployment.ID)
				err = deploymentModels.New().UpdateCustomDomainStatus(deployment.ID, deploymentModels.CustomDomainStatusActive)
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to mark deployment %s as custom domain confirmed. details: %w", deployment.ID, err))
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

					match, err := checkCustomDomain(cd.Domain, cd.Secret)
					if err != nil {
						log.Printf("failed to check custom domain %s for vm %s. details: %v\n", cd.Domain, vm.ID, err)
					}

					if !match {
						err = vmModels.New().UpdateCustomDomainStatus(vm.ID, portName, vmModels.CustomDomainStatusVerificationFailed)
						if err != nil {
							utils.PrettyPrintError(fmt.Errorf("custom domain verification failed. details: %w", err))
							continue
						}

						continue
					}

					log.Printf("marking custom domain %s as confirmed for vm %s\n", cd.Domain, vm.ID)
					err = vmModels.New().UpdateCustomDomainStatus(vm.ID, portName, vmModels.CustomDomainStatusActive)
					if err != nil {
						utils.PrettyPrintError(fmt.Errorf("failed to mark deployment %s as custom domain confirmed. details: %w", vm.ID, err))
						continue
					}
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
		case <-time.After(1 * time.Second):
			workers.ReportStatus("smConfirmer")

		case <-time.After(5 * time.Second):
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

func vmConfirmer(ctx context.Context) {
	defer log.Println("vmConfirmer stopped")

	for {
		select {
		case <-time.After(1 * time.Second):
			workers.ReportStatus("vmConfirmer")

		case <-time.After(5 * time.Second):
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

func checkCustomDomain(domain string, secret string) (bool, error) {
	subDomain := config.Config.Deployment.CustomDomainTxtRecordSubdomain

	txtRecordDomain := subDomain + "." + domain
	txtRecord, err := net.LookupTXT(txtRecordDomain)
	if err != nil {
		// If error is "no such host", it means the DNS record does not exist yet
		var targetErr *net.DNSError
		if ok := errors.As(err, &targetErr); ok && targetErr.IsNotFound {
			log.Printf("no TXT record found under %s when confirming custom domain %s\n", subDomain, domain)
			return false, nil
		}

		utils.PrettyPrintError(fmt.Errorf("failed to lookup TXT record under %s for custom domain %s. details: %w", subDomain, domain, err))
		return false, err
	}

	if len(txtRecord) == 0 {
		log.Printf("no TXT record found under %s when confirming custom domain %s\n", subDomain, domain)
		return false, nil
	}

	match := false
	for _, r := range txtRecord {
		if r == secret {
			match = true
			break
		}
	}

	if !match {
		log.Printf("TXT record found under %s for custom domain %s but secret does not match\n", subDomain, domain)
		return false, nil
	}

	return match, nil
}

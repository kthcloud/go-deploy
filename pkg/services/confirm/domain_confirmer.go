package confirm

import (
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/db/resources/deployment_repo"
	"github.com/kthcloud/go-deploy/pkg/db/resources/vm_repo"
	"github.com/kthcloud/go-deploy/pkg/log"
)

// CustomDomainConfirmer is a worker that confirms custom domain setup.
func CustomDomainConfirmer() error {
	subDomain := config.Config.Deployment.CustomDomainTxtRecordSubdomain

	deploymentsWithPendingCustomDomain, err := deployment_repo.New().WithPendingCustomDomain().List()
	if err != nil {
		return err
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
			return err
		}

		if !exists {
			log.Printf("No TXT record found under %s when confirming custom domain %s for deployment %s\n", subDomain, cd.Domain, deployment.ID)
			continue
		}

		if !match {
			received := txtRecord
			expected := cd.Secret
			if len(received) > len(expected) {
				received = received[:len(expected)] + "..."
			}

			log.Printf("TXT record found under %s but secret does not match when confirming custom domain %s for deployment %s (received: %s, expected: %s)\n", subDomain, cd.Domain, deployment.ID, received, expected)
			err = deployment_repo.New().UpdateCustomDomainStatus(deployment.ID, model.CustomDomainStatusVerificationFailed)
			if err != nil {
				return nil
			}

			continue
		}

		log.Printf("Marking custom domain %s as confirmed for deployment %s", cd.Domain, deployment.ID)
		err = deployment_repo.New().UpdateCustomDomainStatus(deployment.ID, model.CustomDomainStatusActive)
		if err != nil {
			return nil
		}
	}

	vmsWithPendingCustomDomain, err := vm_repo.New().ListWithAnyPendingCustomDomain()
	if err != nil {
		return nil
	}

	for _, vm := range vmsWithPendingCustomDomain {
		// Check if user has updated the DNS record with the custom domain secret
		// If yes, mark the deployment as custom domain confirmed
		for portName, port := range vm.PortMap {
			if port.HttpProxy == nil || port.HttpProxy.CustomDomain == nil || port.HttpProxy.CustomDomain.Status == model.CustomDomainStatusActive {
				continue
			}

			cd := port.HttpProxy.CustomDomain
			if cd == nil {
				continue
			}

			exists, match, txtRecord, err := checkCustomDomain(cd.Domain, cd.Secret)
			if err != nil {
				return nil
			}

			if !exists {
				log.Printf("No TXT record found under %s when confirming custom domain %s for vm %s\n", subDomain, cd.Domain, vm.ID)
				continue
			}

			if !match {
				received := txtRecord
				expected := cd.Secret
				if len(received) > len(expected) {
					received = received[:len(expected)] + "..."
				}

				log.Printf("TXT record found under %s but secret does not match when confirming custom domain %s for vm %s (received: %s, expected: %s)\n", subDomain, cd.Domain, vm.ID, received, expected)
				err = vm_repo.New().UpdateCustomDomainStatus(vm.ID, portName, model.CustomDomainStatusVerificationFailed)
				if err != nil {
					return err
				}

				continue
			}

			log.Printf("Marking custom domain %s as confirmed for vm %s", cd.Domain, vm.ID)
			err = vm_repo.New().UpdateCustomDomainStatus(vm.ID, portName, model.CustomDomainStatusActive)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

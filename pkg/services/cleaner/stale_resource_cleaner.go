package cleaner

import (
	bodyV1 "go-deploy/dto/v1/body"
	bodyV2 "go-deploy/dto/v2/body"
	"go-deploy/models/model"
	"go-deploy/models/version"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/deployment_repo"
	"go-deploy/pkg/db/resources/vm_repo"
	"go-deploy/pkg/log"
	"go-deploy/service"
	"time"
)

func staleResourceCleaner() error {
	// Fetch all resources that have not been accessed in 3 months and disable them
	deployments, err := deployment_repo.New().LastAccessedBefore(time.Now().Add(-config.Config.Deployment.Lifetime)).List()
	if err != nil {
		return err
	}

	for _, deployment := range deployments {
		if deployment.GetMainApp().Replicas == 0 {
			continue
		}

		zone := config.Config.GetZone(deployment.Zone)
		if zone == nil {
			log.Printf("Zone %s not found", deployment.Zone)
			continue
		}

		if !zone.Enabled {
			continue
		}

		log.Printf("Disabling deployment %s due to inactivity", deployment.Name)

		// Set its replicas to 0
		replicas := 0
		err = service.V1().Deployments().Update(deployment.ID, &bodyV1.DeploymentUpdate{
			Replicas: &replicas,
		})
		if err != nil {
			return err
		}
	}

	vms, err := vm_repo.New(version.V2).LastAccessedBefore(time.Now().Add(-config.Config.VM.Lifetime)).List()
	if err != nil {
		return err
	}

	for _, vm := range vms {
		if k8sVM := vm.Subsystems.K8s.GetVM(vm.Name); k8sVM != nil && !k8sVM.Running {
			continue
		}

		log.Printf("Disabling VM %s due to inactivity", vm.Name)

		// Stop the VM
		err = service.V2().VMs().DoAction(vm.ID, &bodyV2.VmActionCreate{
			Action: model.ActionStop,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

package cleaner

import (
	"time"

	bodyV2 "github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/models/version"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/db/resources/deployment_repo"
	"github.com/kthcloud/go-deploy/pkg/db/resources/vm_repo"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/subsystems"
	"github.com/kthcloud/go-deploy/service"
)

func staleResourceCleaner() error {
	// Fetch all resources that have not been accessed in 3 months and disable them
	deployments, err := deployment_repo.New().LastAccessedBefore(time.Now().Add(-config.Config.Deployment.Lifetime)).List()
	if err != nil {
		return err
	}

	for _, deployment := range deployments {
		if deployment.NeverStale || deployment.GetMainApp().Replicas == 0 {
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
		err = service.V2().Deployments().Update(deployment.ID, &bodyV2.DeploymentUpdate{
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
		if k8sVM := &vm.Subsystems.K8s.VM; vm.NeverStale || subsystems.NotCreated(k8sVM) || !k8sVM.Running {
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

package status_update

import (
	"go-deploy/models/model"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/vm_repo"
	"go-deploy/service"
)

func VmStatusFetcher() error {
	// So we can list all the statuses in the cluster and match them with the Deployments
	var allVmStatuses []model.VmStatus
	var allVmiStatuses []model.VmiStatus

	for _, zone := range config.Config.EnabledZones() {
		z := zone
		vmStatuses, err := service.V2().VMs().K8s().ListVmStatus(&z)
		if err != nil {
			return err
		}

		vmiStatuses, err := service.V2().VMs().K8s().ListVmiStatus(&z)
		if err != nil {
			return err
		}

		allVmStatuses = append(allVmStatuses, vmStatuses...)
		allVmiStatuses = append(allVmiStatuses, vmiStatuses...)
	}

	vrc := vm_repo.New()

	for _, status := range allVmStatuses {
		err := vrc.SetStatusByName(status.Name, parseVmStatus(&status))
		if err != nil {
			return err
		}
	}

	for _, status := range allVmiStatuses {
		if status.Host == nil {
			err := vrc.UnsetCurrentHost(status.Name)
			if err != nil {
				return err
			}
		} else {
			err := vrc.SetCurrentHost(status.Name, &model.VmHost{Name: *status.Host})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

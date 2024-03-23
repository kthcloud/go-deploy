package migrator

import (
	"fmt"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/deployment_repo"
	"go-deploy/pkg/db/resources/sm_repo"
	"go-deploy/pkg/db/resources/vm_repo"
	k8sServiceDeployment "go-deploy/service/v1/deployments/k8s_service"
	k8sServiceSM "go-deploy/service/v1/sms/k8s_service"
	k8sServiceVM "go-deploy/service/v1/vms/k8s_service"
)

// Migrate will run as early as possible in the program, and it will never be called again.
func Migrate() error {
	migrations := getMigrations()

	if len(migrations) > 0 {
		for name, migration := range migrations {
			fmt.Printf("- %s (%d/%d)\n", name, 1, len(migrations))
			if err := migration(); err != nil {
				return fmt.Errorf("migration %s failed. details: %w", name, err)
			}
		}
	} else {
		fmt.Println("No migrations to run")
	}

	return nil
}

// getMigrations returns a map of migrations to run.
// add a migration to the list of functions to run.
// clear when prod has run it once.
//
// the migrations must be **idempotent**.
//
// add a date to the migration name to make it easier to identify.
func getMigrations() map[string]func() error {
	return map[string]func() error{
		"moveIntoNewDeployNamespaces_2024_02_16": moveIntoNewDeployNamespaces,
	}
}

func moveIntoNewDeployNamespaces() error {
	deployments, err := deployment_repo.New().List()
	if err != nil {
		return err
	}

	// Migrate deployments to new namespaces
	for _, deployment := range deployments {
		zone := config.Config.Deployment.GetZone(deployment.Zone)
		if zone == nil {
			return fmt.Errorf("zone %s not found", deployment.Zone)
		}

		if zone.Namespaces.Deployment == "" {
			return fmt.Errorf("namespace not found for zone %s", deployment.Zone)
		}

		client := k8sServiceDeployment.New(nil)

		if deployment.Subsystems.K8s.Namespace.Name != "" && deployment.Subsystems.K8s.Namespace.Name != zone.Namespaces.Deployment {
			// Delete Namespace
			_, kc, _, err := client.Get(k8sServiceDeployment.OptsNoGenerator(deployment.ID))
			if err != nil {
				return err
			}

			// Deleting namespace will cause a cascade delete of all resources in the namespace
			if err := kc.DeleteNamespace(deployment.Subsystems.K8s.Namespace.Name); err != nil {
				return err
			}

			// Finally, trigger a repair to recreate the resources in the new namespace with the correct name
			err = client.Repair(deployment.ID)
			if err != nil {
				return err
			}
		}
	}

	// Migrate VM (http proxies) to new namespaces
	vms, err := vm_repo.New().List()
	if err != nil {
		return err
	}

	for _, vm := range vms {
		if vm.DeploymentZone == nil {
			continue
		}

		zone := config.Config.Deployment.GetZone(*vm.DeploymentZone)
		if zone == nil {
			return fmt.Errorf("zone %s not found", vm.Zone)
		}

		if zone.Namespaces.VM == "" {
			return fmt.Errorf("namespace not found for zone %s", vm.Zone)
		}

		client := k8sServiceVM.New(nil)

		if vm.Subsystems.K8s.Namespace.Name != "" && vm.Subsystems.K8s.Namespace.Name != zone.Namespaces.VM {
			// Delete Namespace
			_, kc, _, err := client.Get(k8sServiceVM.OptsNoGenerator(vm.ID))
			if err != nil {
				return err
			}

			// Deleting namespace will cause a cascade delete of all resources in the namespace
			if err := kc.DeleteNamespace(vm.Subsystems.K8s.Namespace.Name); err != nil {
				return err
			}

			// Finally, trigger a repair to recreate the resources in the new namespace with the correct name
			err = client.Repair(vm.ID)
			if err != nil {
				return err
			}
		}
	}

	// Migrate SMs to new namespaces
	sms, err := sm_repo.New().List()
	if err != nil {
		return err
	}

	for _, sm := range sms {
		zone := config.Config.Deployment.GetZone(sm.Zone)
		if zone == nil {
			return fmt.Errorf("zone not found")
		}

		if zone.Namespaces.System == "" {
			return fmt.Errorf("namespace not found for zone")
		}

		client := k8sServiceSM.New(nil)

		if sm.Subsystems.K8s.Namespace.Name != "" && sm.Subsystems.K8s.Namespace.Name != zone.Namespaces.System {
			// Delete Namespace
			_, kc, _, err := client.Get(k8sServiceSM.OptsNoGenerator(sm.ID))
			if err != nil {
				return err
			}

			// Deleting namespace will cause a cascade delete of all resources in the namespace
			if err := kc.DeleteNamespace(sm.Subsystems.K8s.Namespace.Name); err != nil {
				return err
			}

			// Finally, trigger a repair to recreate the resources in the new namespace with the correct name
			err = client.Repair(sm.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

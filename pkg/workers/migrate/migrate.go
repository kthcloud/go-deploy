package migrator

import (
	"fmt"
	deploymentModels "go-deploy/models/sys/deployment"
	smModels "go-deploy/models/sys/sm"
	vmModels "go-deploy/models/sys/vm"
	"go-deploy/pkg/config"
	k8sServiceDeployment "go-deploy/service/v1/deployments/k8s_service"
	k8sServiceSM "go-deploy/service/v1/sms/k8s_service"
	k8sServiceVM "go-deploy/service/v1/vms/k8s_service"
	"log"
)

// Migrate will run as early as possible in the program, and it will never be called again.
func Migrate() {
	migrations := getMigrations()

	if len(migrations) > 0 {
		log.Println("migrating...")

		for name, migration := range migrations {
			log.Printf("running migration %s", name)
			if err := migration(); err != nil {
				log.Fatalf("failed to run migration %s. details: %s", name, err)
			}
		}

		log.Println("migrations done")
		return
	}

	log.Println("nothing to migrate")
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
	deployments, err := deploymentModels.New().List()
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

			// Also delete non-namespaced resources
			for _, pv := range deployment.Subsystems.K8s.PvMap {
				if err := kc.DeletePV(pv.Name); err != nil {
					return err
				}
			}
		}

		// Finally, trigger a repair to recreate the resources in the new namespace with the correct name
		err = client.Repair(deployment.ID)
		if err != nil {
			return err
		}
	}

	// Migrate VM (http proxies) to new namespaces
	vms, err := vmModels.New().List()
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

			// Also delete non-namespaced resources
			for _, pv := range vm.Subsystems.K8s.PvMap {
				if err := kc.DeletePV(pv.Name); err != nil {
					return err
				}
			}
		}

		// Finally, trigger a repair to recreate the resources in the new namespace with the correct name
		err = client.Repair(vm.ID)
		if err != nil {
			return err
		}
	}

	// Migrate SMs to new namespaces
	sms, err := smModels.New().List()
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

			// Also delete non-namespaced resources
			for _, pv := range sm.Subsystems.K8s.PvMap {
				if err := kc.DeletePV(pv.Name); err != nil {
					return err
				}
			}
		}

		// Finally, trigger a repair to recreate the resources in the new namespace with the correct name
		err = client.Repair(sm.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

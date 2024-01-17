package migrator

import (
	deploymentModels "go-deploy/models/sys/deployment"
	smModels "go-deploy/models/sys/sm"
	vmModels "go-deploy/models/sys/vm"
	"go-deploy/pkg/subsystems/k8s/keys"
	"go.mongodb.org/mongo-driver/bson"
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
		"migrateK8sServiceSelector_2024-01-17": migrateK8sServiceSelector_2024_01_17,
	}
}

func migrateK8sServiceSelector_2024_01_17() error {
	deployments, err := deploymentModels.New().List()
	if err != nil {
		return err
	}

	for _, deployment := range deployments {
		if len(deployment.Subsystems.K8s.ServiceMap) == 0 {
			continue
		}

		for mapName, s := range deployment.Subsystems.K8s.ServiceMap {
			if len(s.Selector) == 0 {
				s.Selector = map[string]string{
					keys.ManifestLabelName: s.Name,
				}

				deployment.Subsystems.K8s.ServiceMap[mapName] = s
			}
		}

		err = deploymentModels.New().SetWithBsonByID(deployment.ID, bson.D{
			{"subsystems.k8s.serviceMap", deployment.Subsystems.K8s.ServiceMap},
		})
		if err != nil {
			return err
		}
	}

	vms, err := vmModels.New().List()
	if err != nil {
		return err
	}

	for _, vm := range vms {
		if len(vm.Subsystems.K8s.ServiceMap) == 0 {
			continue
		}

		for mapName, s := range vm.Subsystems.K8s.ServiceMap {
			if len(s.Selector) == 0 {
				s.Selector = map[string]string{
					keys.ManifestLabelName: s.Name,
				}

				vm.Subsystems.K8s.ServiceMap[mapName] = s
			}
		}

		err = vmModels.New().SetWithBsonByID(vm.ID, bson.D{
			{"subsystems.k8s.serviceMap", vm.Subsystems.K8s.ServiceMap},
		})
		if err != nil {
			return err
		}
	}

	sms, err := smModels.New().List()
	if err != nil {
		return err
	}

	for _, sm := range sms {
		if len(sm.Subsystems.K8s.ServiceMap) == 0 {
			continue
		}

		for mapName, s := range sm.Subsystems.K8s.ServiceMap {
			if len(s.Selector) == 0 {
				s.Selector = map[string]string{
					keys.ManifestLabelName: s.Name,
				}

				sm.Subsystems.K8s.ServiceMap[mapName] = s
			}
		}

		err = smModels.New().SetWithBsonByID(sm.ID, bson.D{
			{"subsystems.k8s.serviceMap", sm.Subsystems.K8s.ServiceMap},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

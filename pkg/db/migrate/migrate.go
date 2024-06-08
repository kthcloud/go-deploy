package migrator

import (
	"fmt"
	"go-deploy/models/version"
	"go-deploy/pkg/db/resources/vm_repo"
	"go-deploy/pkg/log"
	"go-deploy/pkg/subsystems/k8s/models"
	"go.mongodb.org/mongo-driver/bson"
)

// Migrate runs every migration specified in the getMigrations function.
// It should be run as early as possible in the program, and should never be called again.
func Migrate() error {
	migrations := getMigrations()

	if len(migrations) > 0 {
		for name, migration := range migrations {
			log.Printf("- %s (%d/%d)\n", name, 1, len(migrations))
			if err := migration(); err != nil {
				return fmt.Errorf("migration %s failed. details: %w", name, err)
			}
		}
	} else {
		log.Println("No migrations to run")
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
		"moveKubeVirtVmMapToSingleVM": moveKubeVirtVmMapToSingleVM,
	}
}

func moveKubeVirtVmMapToSingleVM() error {
	vms, err := vm_repo.New(version.V2).List()
	if err != nil {
		return fmt.Errorf("failed to list vms: %w", err)
	}

	for _, vm := range vms {
		if vm.Subsystems.K8s.VM.ID != "" {
			continue
		}

		var k8sVM *models.VmPublic
		for _, k8sVmInMap := range vm.Subsystems.K8s.VmMap {
			k8sVM = &k8sVmInMap
			break
		}

		err = vm_repo.New(version.V2).SetWithBsonByID(vm.ID, bson.D{
			{"subsystems.k8s.vm", k8sVM},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

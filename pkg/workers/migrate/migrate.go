package migrator

import (
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/subsystems/cs/models"
	"go-deploy/service/vm_service/cs_service"
	"log"
	"strings"
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
		"remove-old-system-snapshots_2023-11-10": removeOldSystemSnapshots,
	}
}

func removeOldSystemSnapshots() error {
	vms, err := vmModel.New().ListAll()
	if err != nil {
		return err
	}

	for _, vm := range vms {
		snapshots := vm.Subsystems.CS.SnapshotMap

		for _, snapshot := range snapshots {
			if snapshot.UserCreated() || goodSnapshotName(&snapshot) {
				continue
			}

			if err = cs_service.DeleteSnapshot(vm.ID, snapshot.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func goodSnapshotName(snapshot *models.SnapshotPublic) bool {
	allowed := []string{"daily", "weekly", "monthly"}
	for _, name := range allowed {
		if strings.Contains(snapshot.Name, name) {
			return true
		}
	}

	return false
}

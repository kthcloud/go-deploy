package migrator

import (
	vmModels "go-deploy/models/sys/vm"
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
		"migrateSpecIntoCsVm_2024_01_09": migrateSpecIntoCsVm_2024_01_09,
	}
}

func migrateSpecIntoCsVm_2024_01_09() error {
	vms, err := vmModels.New().List()
	if err != nil {
		return err
	}

	for _, vm := range vms {
		csVM := &vm.Subsystems.CS.VM
		if csVM.CpuCores != 0 && csVM.RAM != 0 {
			continue
		}

		err = vmModels.New().SetWithBsonByID(vm.ID, bson.D{
			{"subsystems.cs.vm.cpuCores", vm.Specs.CpuCores},
			{"subsystems.cs.vm.ram", vm.Specs.RAM},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

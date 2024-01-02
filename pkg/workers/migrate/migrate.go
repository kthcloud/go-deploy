package migrator

import (
	"errors"
	vmModels "go-deploy/models/sys/vm"
	vmPortModels "go-deploy/models/sys/vm_port"
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
		"leaseVmPortsFromOldSystem_2023_10_30": leaseVmPortsFromOldSystem,
	}
}

func leaseVmPortsFromOldSystem() error {
	vms, err := vmModels.New().List()
	if err != nil {
		return err
	}

	for _, vm := range vms {
		for _, pfr := range vm.Subsystems.CS.PortForwardingRuleMap {
			vmPort, err := vmPortModels.New().GetByLease(vm.ID, pfr.PrivatePort)
			if err != nil {
				return err
			}

			if vmPort == nil {

				_, err = vmPortModels.New().Lease(pfr.PublicPort, pfr.PrivatePort, vm.ID, vm.Zone)
				if err != nil {
					if errors.Is(err, vmPortModels.PortNotFoundErr) {
						return err
					}

					return err
				}
			}
		}
	}

	return nil
}

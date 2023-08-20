package migrator

import (
	deploymentModel "go-deploy/models/sys/deployment"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/conf"
	"go.mongodb.org/mongo-driver/bson"
	log "log"
)

// This file is edited when an update in the database schema has occurred.
// Thus, every migration needed will be done programmatically.
// Once a migration is done, clear the file.

// Migrate will  run as early as possible in the program, and it will never be called again.
func Migrate() {

	log.Println("running migrations for deployments and vms without a ZoneID")

	deployments, err := deploymentModel.GetAll()
	if err != nil {
		panic(err)
	}

	vms, err := vmModel.GetAll()
	if err != nil {
		panic(err)
	}

	zone := conf.Env.VM.GetZone("Flemingsberg")
	if zone == nil {
		panic("zone with name Flemingsberg not found")
	}

	migratedDeployments := 0
	for _, deployment := range deployments {
		// set deployment.ZoneID to

		if deployment.Zone != "" {
			continue
		}

		deployment.Zone = zone.Name

		err := deploymentModel.UpdateByName(deployment.Name, bson.D{
			{"zone", zone.Name},
		})
		if err != nil {
			panic(err)
		}

		migratedDeployments++
	}

	migratedVms := 0
	for _, vm := range vms {
		// set vm.ZoneID and vm.Subsystems.CS.VM.ZoneID to vm.ZoneID

		if vm.Zone != "" {
			continue
		}

		vm.Zone = zone.Name

		err := vmModel.UpdateByName(vm.Name, bson.D{
			{"zone", zone.Name},
		})

		if err != nil {
			panic(err)
		}

		migratedVms++
	}

	log.Printf("migrated %d/%d deployments and %d/%d vms", migratedDeployments, len(deployments), migratedVms, len(vms))
}

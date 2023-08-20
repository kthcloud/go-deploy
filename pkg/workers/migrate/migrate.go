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
	log.Println("running migrations for deployments and vms without a zone")

	deployments, err := deploymentModel.GetAll()
	if err != nil {
		panic(err)
	}

	vms, err := vmModel.GetAll()
	if err != nil {
		panic(err)
	}

	deploymentZone := conf.Env.Deployment.GetZone("se-flem")
	if deploymentZone == nil {
		panic("deployment zone with name se-flem not found")
	}
	
	vmZone := conf.Env.VM.GetZone("se-flem")
	if vmZone == nil {
		panic("zone with name se-flem not found")
	}

	migratedDeployments := 0
	for _, deployment := range deployments {
		// set deployment.ZoneID to

		if deployment.Zone != "" {
			continue
		}

		deployment.Zone = deploymentZone.Name

		err := deploymentModel.UpdateByName(deployment.Name, bson.D{
			{"zone", vmZone.Name},
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

		vm.Zone = vmZone.Name

		err := vmModel.UpdateByName(vm.Name, bson.D{
			{"zone", vmZone.Name},
		})

		if err != nil {
			panic(err)
		}

		migratedVms++
	}

	log.Printf("migrated %d/%d deployments and %d/%d vms", migratedDeployments, len(deployments), migratedVms, len(vms))
}

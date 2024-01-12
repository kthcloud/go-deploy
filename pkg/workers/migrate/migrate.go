package migrator

import (
	"go-deploy/models/sys/base/resource"
	deploymentModels "go-deploy/models/sys/deployment"
	vmModels "go-deploy/models/sys/vm"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"strconv"
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
		"migrateCustomDomainStatusReadyToActive_2024_01_12":               migrateCustomDomainStatusReadyToActive_2024_01_12,
		"migrateOldCustomDomainToNewStruct_2024_01_02":                    migrateOldCustomDomainToNewStruct_2024_01_02,
		"migrateAwayFromNullFields_VmHttpProxyAndCustomDomain_2024_01_12": migrateAwayFromNullFields_VmHttpProxyAndCustomDomain_2024_01_12,
		"migratePortListToPortMap_2024_01_12":                             migratePortListToPortMap_2024_01_12,
	}
}

func migrateCustomDomainStatusReadyToActive_2024_01_12() error {
	deployments, err := deploymentModels.New().List()
	if err != nil {
		return err
	}

	for _, deployment := range deployments {
		app := deployment.GetMainApp()
		if app.CustomDomain != nil && app.CustomDomain.Status == "ready" {
			err = deploymentModels.New().SetWithBsonByID(deployment.ID, bson.D{
				{"apps.main.customDomain.status", deploymentModels.CustomDomainStatusActive},
			})

			if err != nil {
				return err
			}
		}
	}

	return nil
}

type oldVM1 struct {
	ID    string `bson:"id"`
	Ports []struct {
		Name      string `bson:"name"`
		HttpProxy *struct {
			CustomDomain *string `bson:"customDomain"`
		}
	}
}

func migrateOldCustomDomainToNewStruct_2024_01_02() error {
	rc := resource.ResourceClient[oldVM1]{
		Collection: vmModels.New().Collection,
	}

	ids, err := vmModels.New().ListIDs()
	if err != nil {
		return err
	}

	for _, id := range ids {
		vm, err := rc.GetByID(id.ID)
		if err != nil {
			if strings.Contains(err.Error(), "error decoding key") {
				continue
			}

			return err
		}

		for idx, port := range vm.Ports {
			if port.HttpProxy != nil && port.HttpProxy.CustomDomain != nil {
				customDomain := vmModels.CustomDomain{
					Domain: *port.HttpProxy.CustomDomain,
					Secret: "",
					Status: deploymentModels.CustomDomainStatusActive,
				}

				updatePath := "ports." + strconv.Itoa(idx) + ".httpProxy.customDomain"

				err = rc.SetWithBsonByID(id.ID, bson.D{
					{updatePath, customDomain},
				})

				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func migratePortListToPortMap_2024_01_12() error {
	vms, err := vmModels.New().List()
	if err != nil {
		return err
	}

	for _, vm := range vms {
		// ssh will always be there, so this check is safe
		if len(vm.PortMap) > 0 {
			continue
		}

		portMap := make(map[string]vmModels.Port)

		for _, port := range vm.Ports_ {
			portMap[port.Name] = port
		}

		err = vmModels.New().SetWithBsonByID(vm.ID, bson.D{
			{"portMap", portMap},
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func migrateAwayFromNullFields_VmHttpProxyAndCustomDomain_2024_01_12() error {
	vms, err := vmModels.New().List()
	if err != nil {
		return err
	}

	for _, vm := range vms {
		for mapName, port := range vm.PortMap {
			if port.HttpProxy == nil {
				err = vmModels.New().UnsetWithBsonByID(vm.ID, "portMap."+mapName+".httpProxy")
				if err != nil {
					return err
				}

				continue
			}

			if port.HttpProxy.CustomDomain == nil {
				err = vmModels.New().UnsetWithBsonByID(vm.ID, "portMap."+mapName+".httpProxy.customDomain")
				if err != nil {
					return err
				}

				continue
			}
		}
	}

	return nil
}

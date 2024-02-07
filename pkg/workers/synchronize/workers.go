package synchronize

import (
	"fmt"
	gpuModels "go-deploy/models/sys/gpu"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/landing"
	"go-deploy/pkg/subsystems/landing/models"
	"go-deploy/pkg/workers"
	"go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"strings"
	"time"
)

func gpuSynchronizer() {
	defer workers.OnStop("gpuSynchronizer")

	reportTick := time.Tick(1 * time.Second)
	tick := time.Tick(3 * time.Second)

	makeError := func(err error) error {
		return fmt.Errorf("failed to synchronize gpus: %w", err)
	}

	for {
		select {
		case <-reportTick:
			workers.ReportUp("gpuSynchronizer")
		case <-tick:

			// Fetch GPUs
			gpuInfo, err := fetchGPUs()
			if err != nil {
				utils.PrettyPrintError(makeError(err))
				continue
			}

			ids := make([]string, 0)
			for _, host := range gpuInfo.GpuInfo.Hosts {
				for _, gpu := range host.GPUs {
					ids = append(ids, createGpuID(host.Name, gpu.Name, gpu.Slot))
				}
			}

			// Delete GPUs without a lease to sync with the landing page
			err = gpuModels.New().WithoutLease().ExcludeIDs(ids...).Erase()
			if err != nil {
				utils.PrettyPrintError(makeError(err))
				continue
			}

			// Update stale GPUs
			err = gpuModels.New().ExcludeIDs(ids...).SetWithBSON(bson.D{{"stale", true}})
			if err != nil {
				utils.PrettyPrintError(makeError(err))
				continue
			}

			// Warn if there are any stale GPUs
			staleGPUs, err := gpuModels.New().WithStale().List()
			if err != nil {
				utils.PrettyPrintError(makeError(err))
				continue
			}

			if len(staleGPUs) > 0 {
				printString := "Stale GPUs detected: \n"
				for _, gpu := range staleGPUs {
					printString += "\t- " + gpu.ID + "\n"
				}
				printString = strings.TrimSuffix(printString, ", ")
				printString += "Detach them from the VMs to prevent unexpected behavior."
				log.Println(printString)
			}

			// Add GPUs to the database
			for _, host := range gpuInfo.GpuInfo.Hosts {
				for _, gpu := range host.GPUs {
					gpuID := createGpuID(host.Name, gpu.Name, gpu.Slot)
					exists, err := gpuModels.New().ExistsByID(gpuID)
					if err != nil {
						utils.PrettyPrintError(fmt.Errorf("failed to fetch gpu by id. details: %w", err))
						continue
					}

					if exists {
						continue
					}

					zone := config.Config.VM.GetZoneByID(host.ZoneID)
					if zone == nil {
						log.Println("GPU zone", host.ZoneID, "not found. Skipping GPU", gpuID)
						continue
					}

					err = gpuModels.New().Create(gpuID, host.Name, gpuModels.GpuData{
						Name:     gpu.Name,
						Slot:     gpu.Slot,
						Vendor:   gpu.Vendor,
						VendorID: gpu.VendorID,
						Bus:      gpu.Bus,
						DeviceID: gpu.DeviceID,
					}, zone.Name)

					if err != nil {
						utils.PrettyPrintError(makeError(err))
						continue
					}
				}
			}
		}
	}
}

func fetchGPUs() (*models.GpuInfoRead, error) {
	makeError := func(err error) error {
		return fmt.Errorf("error fetching gpus: %w", err)
	}

	client, err := landing.New(&landing.ClientConf{
		URL:      config.Config.Landing.URL,
		Username: config.Config.Landing.User,
		Password: config.Config.Landing.Password,

		OidcProvider: config.Config.Keycloak.Url,
		OidcClientID: config.Config.Landing.ClientID,
		OidcRealm:    config.Config.Keycloak.Realm,
	})

	if err != nil {
		return nil, makeError(err)
	}

	gpuInfo, err := client.ReadGpuInfo()
	if err != nil {
		return nil, makeError(err)
	}

	return gpuInfo, nil
}

func deleteRemovedUnleasedGPUs() error {
	err := gpuModels.New().WithoutLease().Delete()
	if err != nil {
		return fmt.Errorf("failed to delete gpu: %w", err)
	}

	return nil
}

func createGpuID(host, gpuName, slot string) string {
	gpuName = strings.Replace(gpuName, " ", "_", -1)
	return fmt.Sprintf("%s-%s-%s", host, gpuName, slot)
}

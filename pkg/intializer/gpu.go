package intializer

import (
	"fmt"
	gpuModel "go-deploy/models/sys/vm/gpu"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/landing"
	"go-deploy/utils"
	"log"
	"strings"
)

func SynchronizeGPUs() {
	client, err := landing.New(&landing.ClientConf{
		URL:      config.Config.Landing.URL,
		Username: config.Config.Landing.User,
		Password: config.Config.Landing.Password,

		OidcProvider: config.Config.Keycloak.Url,
		OidcClientID: config.Config.Landing.ClientID,
		OidcRealm:    config.Config.Keycloak.Realm,
	})

	if err != nil {
		log.Fatal(err)
	}

	gpuInfo, err := client.ReadGpuInfo()
	if err != nil {
		log.Fatal(err)
	}

	configured := 0
	currentGPUs, err := gpuModel.New().GetAll()
	if err != nil {
		log.Fatal(err)
	}

	for _, gpu := range currentGPUs {
		// clear non-leased gpus
		// this is safe since an admin must delete the lease before removing the gpu from the system
		if gpu.Lease.VmID == "" {
			err = gpuModel.New().Delete(gpu.ID)
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to delete gpu. details: %w", err))
			}
		}
	}

	for _, host := range gpuInfo.GpuInfo.Hosts {
		for _, gpu := range host.GPUs {

			id := createGpuID(host.Name, gpu.Name, gpu.Slot)

			current, err := gpuModel.New().GetByID(id)
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to fetch gpu by id. details: %w", err))
				continue
			}

			if current != nil {
				configured++
				continue
			}

			zone := config.Config.VM.GetZoneByID(host.ZoneID)
			if zone == nil {
				log.Println("graphics card zone not found. zone id: ", host.ZoneID, ". skipping gpu: ", id)
				continue
			}

			err = gpuModel.New().Create(id, host.Name, gpuModel.GpuData{
				Name:     gpu.Name,
				Slot:     gpu.Slot,
				Vendor:   gpu.Vendor,
				VendorID: gpu.VendorID,
				Bus:      gpu.Bus,
				DeviceID: gpu.DeviceID,
			}, zone.Name)

			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to create gpu. details: %w", err))
			}

			configured++
		}
	}

	log.Println("configured", configured, "gpus")
}

func createGpuID(host, gpuName, slot string) string {
	gpuName = strings.Replace(gpuName, " ", "_", -1)
	return fmt.Sprintf("%s-%s-%s", host, gpuName, slot)
}

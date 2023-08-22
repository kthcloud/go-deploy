package intializer

import (
	"fmt"
	gpuModel "go-deploy/models/sys/vm/gpu"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/landing"
	"log"
	"strings"
)

func SynchronizeGPUs() {
	client, err := landing.New(&landing.ClientConf{
		URL:      conf.Env.Landing.Url,
		Username: conf.Env.Landing.User,
		Password: conf.Env.Landing.Password,

		OidcProvider: conf.Env.Keycloak.Url,
		OidcClientID: conf.Env.Landing.ClientID,
		OidcRealm:    conf.Env.Keycloak.Realm,
	})

	if err != nil {
		log.Fatal(err)
	}

	gpuInfo, err := client.ReadGpuInfo()
	if err != nil {
		log.Fatal(err)
	}

	configured := 0
	currentGPUs, err := gpuModel.GetAll(nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	for _, gpu := range currentGPUs {
		// clear non-leased gpus
		// this is safe since an admin must delete the lease before removing the gpu from the system
		if gpu.Lease.VmID == "" {
			err = gpuModel.Delete(gpu.ID)
			if err != nil {
				log.Println("failed to delete gpu. details: ", err)
			}
		}
	}

	for _, host := range gpuInfo.GpuInfo.Hosts {
		for _, gpu := range host.GPUs {

			id := createGpuID(host.Name, gpu.Name, gpu.Slot)

			current, err := gpuModel.GetByID(id)
			if err != nil {
				log.Println("failed to fetch gpu by id. details: ", err)
				continue
			}

			if current != nil {
				configured++
				continue
			}

			zone := conf.Env.VM.GetZoneByID(host.ZoneID)
			if zone == nil {
				log.Println("graphics card zone not found. zone id: ", host.ZoneID, ". skipping gpu: ", id)
				continue
			}

			err = gpuModel.Create(id, host.Name, gpuModel.GpuData{
				Name:     gpu.Name,
				Slot:     gpu.Slot,
				Vendor:   gpu.Vendor,
				VendorID: gpu.VendorID,
				Bus:      gpu.Bus,
				DeviceID: gpu.DeviceID,
			}, zone.Name)

			if err != nil {
				log.Println("failed to create gpu. details: ", err)
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

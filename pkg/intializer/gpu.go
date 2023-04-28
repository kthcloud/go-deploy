package intializer

import (
	"fmt"
	vmModel "go-deploy/models/vm"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/landing"
	"log"
	"strings"
)

func SynchronizeGPU() {
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

	oldGpus := 0
	newGpus := 0

	for _, host := range gpuInfo.GpuInfo.Hosts {
		if len(host.GPUs) == 0 {
			continue
		}

		for _, gpu := range host.GPUs {

			id := createGpuID(host.Name, gpu.Name, gpu.Slot)

			current, err := vmModel.GetGpuByID(id)
			if err != nil {
				log.Println("failed to fetch gpu by id. details: ", err)
				continue
			}

			if current != nil {
				oldGpus++
				continue
			}

			err = vmModel.CreateGPU(id, host.Name, vmModel.GpuData{
				Name:     gpu.Name,
				Slot:     gpu.Slot,
				Vendor:   gpu.Vendor,
				VendorID: gpu.VendorID,
				Bus:      gpu.Bus,
				DeviceID: gpu.DeviceID,
			})

			if err != nil {
				log.Println("failed to create gpu. details: ", err)
			}

			newGpus++
		}
	}

	log.Println("created", newGpus, "new gpus")
	log.Println("found", oldGpus, "already configured gpus")
}

func createGpuID(host, gpuName, slot string) string {
	gpuName = strings.Replace(gpuName, " ", "_", -1)
	return fmt.Sprintf("%s-%s-%s", host, gpuName, slot)
}

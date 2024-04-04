package synchronize

import (
	"fmt"
	"go-deploy/models/model"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/gpu_group_repo"
	"go-deploy/pkg/db/resources/gpu_repo"
	"go-deploy/pkg/log"
	"go-deploy/pkg/subsystems/sys-api"
	"go-deploy/pkg/subsystems/sys-api/models"
	"go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
)

// gpuSynchronizer synchronizes the GPUs in the database with the sys-api page.
// Whenever a GPU is added or removed from a machine, the sys-api is updated, and this
// worker will synchronize the database with the sys-api
func gpuSynchronizer() error {
	// Fetch GPUs
	gpuInfo, err := fetchGPUs()
	if err != nil {
		return err
	}

	// Synchronize GPUs v1
	err = synchronizeGpusV1(gpuInfo)
	if err != nil {
		return err
	}

	// Synchronize GPUs v2
	err = synchronizeGpusV2(gpuInfo)
	if err != nil {
		return err
	}

	return nil
}

func synchronizeGpusV1(gpuInfo *models.GpuInfoRead) error {
	ids := make([]string, 0)
	for _, host := range gpuInfo.GpuInfo.Hosts {
		for _, gpu := range host.GPUs {
			ids = append(ids, createGpuID(host.Name, gpu.Name, gpu.Slot))
		}
	}

	// Delete GPUs without a lease to sync with the sys-api
	err := gpu_repo.New().WithoutLease().ExcludeIDs(ids...).Erase()
	if err != nil {
		return err
	}

	// Update stale GPUs
	err = gpu_repo.New().ExcludeIDs(ids...).SetWithBSON(bson.D{{"stale", true}})
	if err != nil {
		return err
	}

	// Warn if there are any stale GPUs
	staleGPUs, err := gpu_repo.New().WithStale().List()
	if err != nil {
		return err
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
			exists, err := gpu_repo.New().ExistsByID(gpuID)
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to fetch gpu_repo by id. details: %w", err))
				continue
			}

			if exists {
				continue
			}

			zone := config.Config.VM.GetZone(host.Zone)
			if zone == nil {
				log.Println("GPU zone", host.Zone, "not found. Skipping GPU", gpuID)
				continue
			}

			err = gpu_repo.New().Create(gpuID, host.Name, model.GpuData{
				Name:     gpu.Name,
				Slot:     gpu.Slot,
				Vendor:   gpu.Vendor,
				VendorID: gpu.VendorID,
				Bus:      gpu.Bus,
				DeviceID: gpu.DeviceID,
			}, zone.Name)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func synchronizeGpusV2(gpuInfo *models.GpuInfoRead) error {
	// Determine groups
	groups := make(map[string]map[string]model.GpuGroup)
	for _, host := range gpuInfo.GpuInfo.Hosts {
		for _, gpu := range host.GPUs {
			groupName := createGpuGroupName(&gpu)
			if groupName == nil {
				continue
			}
			groupID := utils.HashStringAlphanumericLower(fmt.Sprintf("%s-%s", *groupName, host.Zone))

			// rtx5000: 1eb0  rtx-a6000: 2230
			if groups[host.Zone] == nil {
				groups[host.Zone] = make(map[string]model.GpuGroup)
			}

			if group, ok := groups[host.Zone][groupID]; !ok {
				groups[host.Zone][groupID] = model.GpuGroup{
					Name:        *groupName,
					DisplayName: gpu.Name,
					Zone:        host.Zone,
					Total:       1,
					Vendor:      gpu.Vendor,
					VendorID:    gpu.VendorID,
					DeviceID:    gpu.DeviceID,
				}
			} else {
				group.Total++
				groups[host.Zone][*groupName] = group
			}
		}
	}

	// Update the groups in the database and delete groups that no longer exist
	for zone, groupMap := range groups {
		for _, group := range groupMap {
			exists, err := gpu_group_repo.New().WithZone(zone).ExistsByID(group.ID)
			if err != nil {
				return err
			}

			if !exists {
				err := gpu_group_repo.New().Create(group.Name, group.DisplayName, zone, group.Vendor, group.DeviceID, group.VendorID, group.Total)
				if err != nil {
					return err
				}
			} else {
				err = gpu_group_repo.New().WithZone(zone).SetWithBsonByID(group.ID, bson.D{
					{"name", group.Name},
					{"displayName", group.DisplayName},
					{"total", group.Total},
					{"vendor", group.Vendor},
					{"vendorId", group.VendorID},
					{"deviceId", group.DeviceID},
				})
				if err != nil {
					return err
				}
			}
		}
	}

	// Delete groups that no longer exist
	groupsInDb, err := gpu_group_repo.New().List()
	if err != nil {
		return err
	}

	for _, group := range groupsInDb {
		if _, ok := groups[group.Zone][group.ID]; !ok {
			err := gpu_group_repo.New().WithZone(group.Zone).EraseByID(group.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func fetchGPUs() (*models.GpuInfoRead, error) {
	makeError := func(err error) error {
		return fmt.Errorf("error fetching gpus: %w", err)
	}

	client, err := sys_api.New(&sys_api.ClientConf{
		URL:      config.Config.SysApi.URL,
		Username: config.Config.SysApi.User,
		Password: config.Config.SysApi.Password,

		OidcProvider: config.Config.Keycloak.Url,
		OidcClientID: config.Config.SysApi.ClientID,
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

func createGpuID(host, gpuName, slot string) string {
	gpuName = strings.Replace(gpuName, " ", "_", -1)
	return fmt.Sprintf("%s-%s-%s", host, gpuName, slot)
}

func createGpuGroupName(gpu *models.GPU) *string {
	vendor := strings.ToLower(gpu.Vendor)

	if strings.Contains(vendor, "nvidia") {
		vendor = "nvidia"
	} else {
		// Right now we only support NVIDIA GPUs
		return nil
	}

	device := strings.Replace(strings.ToLower(gpu.Name), " ", "-", -1)

	groupName := fmt.Sprintf("%s/%s", vendor, device)
	return &groupName
}

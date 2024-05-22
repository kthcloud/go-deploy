package synchronize

import (
	"fmt"
	configModels "go-deploy/models/config"
	"go-deploy/models/model"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/gpu_group_repo"
	"go-deploy/pkg/log"
	"go-deploy/pkg/subsystems/sys-api"
	"go-deploy/pkg/subsystems/sys-api/models"
	"go-deploy/service"
	"go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
)

// GpuSynchronizer synchronizes the GPUs in the database with the sys-api page.
// Whenever a GPU is added or removed from a machine, the sys-api is updated, and this
// worker will synchronize the database with the sys-api
func GpuSynchronizer() error {
	// Fetch GPUs
	gpuInfo, err := fetchGPUs()
	if err != nil {
		return err
	}

	// Synchronize GPUs v2
	err = synchronizeGpus(gpuInfo)
	if err != nil {
		return err
	}

	return nil
}

func synchronizeGpus(gpuInfo *models.GpuInfoRead) error {
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

			if group, ok := groups[host.Zone][*groupName]; !ok {
				groups[host.Zone][*groupName] = model.GpuGroup{
					ID:          groupID,
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
		if _, ok := groups[group.Zone][group.Name]; !ok {
			err := gpu_group_repo.New().WithZone(group.Zone).EraseByID(group.ID)
			if err != nil {
				return err
			}
		}
	}

	// Sync GPU groups with backend
	deployV2 := service.V2()

	groupsByZone := make(map[string][]model.GpuGroup)
	for _, group := range groupsInDb {
		groupsByZone[group.Zone] = append(groupsByZone[group.Zone], group)
	}

	for zoneName, groups := range groupsByZone {
		zone := config.Config.GetZone(zoneName)
		if zone == nil {
			log.Println("Zone", zoneName, "not found. Skipping GPU synchronization.")
			continue
		}

		if !zone.HasCapability(configModels.ZoneCapabilityVM) {
			continue
		}

		err = deployV2.VMs().K8s().Synchronize(zoneName, groups)
		if err != nil {
			return err
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

		UseMock: config.Config.SysApi.UseMock,
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

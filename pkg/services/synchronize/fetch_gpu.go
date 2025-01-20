package synchronize

import (
	"fmt"
	"strings"

	"github.com/kthcloud/go-deploy/dto/v2/body"
	configModels "github.com/kthcloud/go-deploy/models/config"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/db/resources/gpu_group_repo"
	"github.com/kthcloud/go-deploy/pkg/db/resources/system_gpu_info_repo"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/service"
	"github.com/kthcloud/go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
)

// GpuSynchronizer synchronizes the GPUs in the database with the sys-api page.
// Whenever a GPU is added or removed from a machine, the sys-api is updated, and this
// worker will synchronize the database with the sys-api
func GpuSynchronizer() error {
	gpuInfo, err := listLatestGPUs()
	if err != nil {
		return err
	}

	if gpuInfo == nil {
		return nil
	}

	err = synchronizeGPUs(gpuInfo)
	if err != nil {
		return err
	}

	return nil
}

func synchronizeGPUs(gpuInfo *body.SystemGpuInfo) error {
	// Determine groups
	groups := make(map[string]map[string]model.GpuGroup)
	for _, host := range gpuInfo.HostGpuInfo {
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
					{Key: "name", Value: group.Name},
					{Key: "displayName", Value: group.DisplayName},
					{Key: "total", Value: group.Total},
					{Key: "vendor", Value: group.Vendor},
					{Key: "vendorId", Value: group.VendorID},
					{Key: "deviceId", Value: group.DeviceID},
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

		if !zone.Enabled {
			continue
		}

		err = deployV2.VMs().K8s().Synchronize(zoneName, groups)
		if err != nil {
			return err
		}
	}

	return nil
}

func listLatestGPUs() (*body.SystemGpuInfo, error) {
	makeError := func(err error) error {
		return fmt.Errorf("error fetching gpus: %w", err)
	}

	systemGpuInfo, err := system_gpu_info_repo.New().List()
	if err != nil {
		return nil, makeError(err)
	}

	var result *body.SystemGpuInfo
	if len(systemGpuInfo) > 0 {
		result = &systemGpuInfo[0].GpuInfo
	}

	if config.Config.GPU.AddMock {
		// Add one mock GPUs in each zone
		for _, zone := range config.Config.EnabledZones() {
			if result == nil {
				result = &body.SystemGpuInfo{}
			}

			result.HostGpuInfo = append(result.HostGpuInfo, body.HostGpuInfo{
				HostBase: body.HostBase{
					Name:        "Mock Host 1",
					DisplayName: "Mock Host 1",
					Zone:        zone.Name,
				},
				GPUs: []body.GpuInfo{{
					Name:        "Mock GPU 1",
					Vendor:      "NVIDIA",
					VendorID:    "10de",
					DeviceID:    "1eb0",
					Passthrough: true,
				}, {
					Name:        "Mock GPU 2",
					Vendor:      "NVIDIA",
					VendorID:    "10de",
					DeviceID:    "2230",
					Passthrough: true,
				}},
			})

			result.HostGpuInfo = append(result.HostGpuInfo, body.HostGpuInfo{
				HostBase: body.HostBase{
					Name:        "Mock Host 2",
					DisplayName: "Mock Host 2",
					Zone:        zone.Name,
				},
				GPUs: []body.GpuInfo{{
					Name:        "Mock GPU 1",
					Vendor:      "NVIDIA",
					VendorID:    "10de",
					DeviceID:    "1eb0",
					Passthrough: true,
				}},
			})
		}
	}

	return result, nil
}

func createGpuGroupName(gpu *body.GpuInfo) *string {
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

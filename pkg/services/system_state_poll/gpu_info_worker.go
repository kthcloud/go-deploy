package system_state_poll

import (
	"fmt"
	"go-deploy/dto/v2/body"
	"go-deploy/models/model"
	"go-deploy/pkg/db/resources/host_repo"
	"go-deploy/pkg/db/resources/system_gpu_info_repo"
	wErrors "go-deploy/pkg/services/errors"
	"go-deploy/pkg/subsystems/host_api"
	"go-deploy/utils"
	"sync"
	"time"
)

func GetHostGpuInfo() ([]body.HostGpuInfo, error) {
	allHosts, err := host_repo.New().Activated().List()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch hosts. details: %s", err)
	}

	outputs := make([]*body.HostGpuInfo, len(allHosts))
	mu := sync.RWMutex{}

	err = ForEachHost("fetch-capacities", allHosts, func(idx int, host *model.Host) error {
		makeError := func(err error) error {
			return fmt.Errorf("failed to get capacities for host %s. details: %s", host.IP, err)
		}

		client := host_api.NewClient(host.ApiURL())

		hostApiGpus, err := client.GetGpuInfo()
		if err != nil {
			return makeError(err)
		}

		var gpuInfo []body.GpuInfo
		for _, gpu := range hostApiGpus {
			// Right now go-deploy only supports GPUs that are passthrough.
			// In the future it could be extended to support non-passthrough GPUs to, for example, allow GPUs in deployments
			if !gpu.Passthrough {
				continue
			}

			gpuInfo = append(gpuInfo, body.GpuInfo{
				Name:        gpu.Name,
				Slot:        gpu.Slot,
				Vendor:      gpu.Vendor,
				VendorID:    gpu.VendorID,
				Bus:         gpu.Bus,
				DeviceID:    gpu.DeviceID,
				Passthrough: gpu.Passthrough,
			})
		}

		hostCapacities := body.HostGpuInfo{
			GPUs: gpuInfo,
			HostBase: body.HostBase{
				Name:        host.Name,
				DisplayName: host.DisplayName,
				Zone:        host.Zone,
			},
		}

		mu.Lock()
		outputs[idx] = &hostCapacities
		mu.Unlock()

		return nil
	})

	return utils.WithoutNils(outputs), err
}

func GpuInfoWorker() error {
	hostGpuInfo, err := GetHostGpuInfo()
	if err != nil {
		return err
	}

	if len(hostGpuInfo) == 0 {
		return wErrors.NoHostsErr
	}

	return system_gpu_info_repo.New(500).Save(&body.TimestampedSystemGpuInfo{
		GpuInfo: body.SystemGpuInfo{
			HostGpuInfo: hostGpuInfo,
		},
		Timestamp: time.Now(),
	})
}

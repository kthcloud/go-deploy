package system_state_poll

import (
	"fmt"
	"go-deploy/dto/v2/body"
	"go-deploy/models/model"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/host_repo"
	"go-deploy/pkg/db/resources/system_capacities_repo"
	"go-deploy/pkg/subsystems/host_api"
	"go-deploy/utils"
	"sync"
	"time"
)

func GetHostCapacities() ([]body.HostCapacities, error) {
	allHosts, err := host_repo.New().Activated().List()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch hosts. details: %s", err)
	}

	outputs := make([]*body.HostCapacities, len(allHosts))
	mu := sync.RWMutex{}

	err = ForEachHost("fetch-capacities", allHosts, func(idx int, host *model.Host) error {
		makeError := func(err error) error {
			return fmt.Errorf("failed to get capacities for host %s. details: %s", host.IP, err)
		}

		client := host_api.NewClient(host.ApiURL())

		hostApiCapacities, err := client.GetCapacities()
		if err != nil {
			return makeError(err)
		}

		hostCapacities := body.HostCapacities{
			CpuCore: body.CpuCoreCapacities{
				Total: hostApiCapacities.CPU.Cores,
			},
			RAM: body.RamCapacities{
				Total: hostApiCapacities.RAM.Total,
			},
			GPU: body.GpuCapacities{
				Total: hostApiCapacities.GPU.Count,
			},
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

func CapacitiesWorker() error {
	// Hosts
	hostCapacities, err := GetHostCapacities()
	if err != nil {
		return err
	}

	if len(hostCapacities) == 0 {
		hostCapacities = make([]body.HostCapacities, 0)
	}

	cpuCoreTotal := 0
	ramTotal := 0
	gpuTotal := 0

	clusters := make([]body.ClusterCapacities, 0)
	// Add empty cluster capacities for each zone
	for _, zone := range config.Config.EnabledZones() {
		clusters = append(clusters, body.ClusterCapacities{
			Name: zone.Name,
			CpuCore: body.CpuCoreCapacities{
				Total: 0,
			},
			RAM: body.RamCapacities{
				Total: 0,
			},
			GPU: body.GpuCapacities{
				Total: 0,
			},
		})
	}

	for _, host := range hostCapacities {
		cpuCoreTotal += host.CpuCore.Total
		ramTotal += host.RAM.Total
		gpuTotal += host.GPU.Total

		// Add host capacities to the corresponding cluster
		for i, cluster := range clusters {
			if cluster.Name == host.Zone {
				clusters[i].CpuCore.Total += host.CpuCore.Total
				clusters[i].RAM.Total += host.RAM.Total
				clusters[i].GPU.Total += host.GPU.Total
				break
			}
		}
	}

	collected := body.SystemCapacities{
		RAM: body.RamCapacities{
			Total: ramTotal,
		},
		CpuCore: body.CpuCoreCapacities{
			Total: cpuCoreTotal,
		},
		GPU: body.GpuCapacities{
			Total: gpuTotal,
		},
		Hosts: hostCapacities,
	}

	return system_capacities_repo.New(500).Save(&body.TimestampedSystemCapacities{
		Capacities: collected,
		Timestamp:  time.Now(),
	})
}

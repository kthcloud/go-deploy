package system_state_poll

import (
	"fmt"
	"go-deploy/dto/v2/body"
	"go-deploy/models/model"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/host_repo"
	"go-deploy/pkg/db/resources/system_capacities_repo"
	wErrors "go-deploy/pkg/services/errors"
	"go-deploy/pkg/subsystems/host_api"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/utils"
	"k8s.io/client-go/kubernetes"
	"log"
	"sync"
	"time"
)

func GetClusterCapacities() ([]body.ClusterCapacities, error) {
	clients := make(map[string]kubernetes.Clientset)
	for _, zone := range config.Config.EnabledZones() {
		if zone.K8s.Client == nil {
			continue
		}

		clients[zone.Name] = *zone.K8s.Client
		break
	}

	outputs := make([]*body.ClusterCapacities, len(clients))
	mu := sync.Mutex{}

	ForEachCluster("fetch-k8s-stats", clients, func(worker int, name string, cluster *kubernetes.Clientset) error {
		makeError := func(err error) error {
			return fmt.Errorf("failed to list pods from cluster %s. details: %s", name, err)
		}

		client, err := k8s.New(&k8s.ClientConf{
			K8sClient: cluster,
		})

		if err != nil {
			log.Println(makeError(err))
			return nil
		}

		nodes, err := client.ListNodes()
		if err != nil {
			log.Println(makeError(err))
			return nil
		}

		clusterCapacities := body.ClusterCapacities{
			Name:    name,
			RAM:     body.RamCapacities{},
			CpuCore: body.CpuCoreCapacities{},
		}

		for _, node := range nodes {
			clusterCapacities.RAM.Total += node.RAM.Total
			clusterCapacities.CpuCore.Total += node.CPU.Total
		}

		mu.Lock()
		outputs[worker] = &clusterCapacities
		mu.Unlock()

		return nil
	})

	return utils.WithoutNils(outputs), nil
}

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

		capacities, err := client.GetCapacities()
		if err != nil {
			return makeError(err)
		}

		hostCapacities := body.HostCapacities{
			Capacities: *capacities,
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
	// Cluster
	clusterCapacities, err := GetClusterCapacities()
	if err != nil {
		return err
	}

	if clusterCapacities == nil {
		clusterCapacities = make([]body.ClusterCapacities, 0)
	}

	// Hosts
	hostCapacities, err := GetHostCapacities()
	if err != nil {
		return err
	}

	if len(hostCapacities) == 0 {
		hostCapacities = make([]body.HostCapacities, 0)
	}

	if len(hostCapacities) == 0 && clusterCapacities == nil {
		return wErrors.NoHostsErr
	}

	gpuTotal := 0
	for _, host := range hostCapacities {
		gpuTotal += host.GPU.Count
	}

	collected := body.SystemCapacities{
		RAM:     body.RamCapacities{},
		CpuCore: body.CpuCoreCapacities{},
		GPU: body.GpuCapacities{
			Total: gpuTotal,
		},
		Hosts: hostCapacities,
	}

	for _, cluster := range clusterCapacities {
		collected.RAM.Total += cluster.RAM.Total
		collected.CpuCore.Total += cluster.CpuCore.Total
	}

	return system_capacities_repo.New().Save(&body.TimestampedSystemCapacities{
		Capacities: collected,
		Timestamp:  time.Now(),
	})
}

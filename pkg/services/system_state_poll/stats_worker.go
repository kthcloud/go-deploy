package system_state_poll

import (
	"fmt"
	"go-deploy/dto/v2/body"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/system_stats_repo"
	wErrors "go-deploy/pkg/services/errors"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/utils"
	"k8s.io/client-go/kubernetes"
	"log"
	"sync"
	"time"
)

func GetClusterStats() ([]body.ClusterStats, error) {
	clients := make(map[string]kubernetes.Clientset)
	for _, zone := range config.Config.EnabledZones() {
		if zone.K8s.Client == nil {
			continue
		}

		clients[zone.Name] = *zone.K8s.Client
		break
	}

	outputs := make([]*body.ClusterStats, len(clients))
	mu := sync.Mutex{}

	ForEachCluster("fetch-k8s-stats", clients, func(worker int, name string, cluster *kubernetes.Clientset) error {
		makeError := func(err error) error {
			return fmt.Errorf("failed to list pods from cluster %s. details: %s", name, err)
		}

		client, err := k8s.New(&k8s.ClientConf{K8sClient: cluster})
		if err != nil {
			log.Println(makeError(err))
			return nil
		}

		pods, err := client.CountPods()
		if err != nil {
			log.Println(makeError(err))
			return nil
		}

		mu.Lock()
		outputs[worker] = &body.ClusterStats{Name: name, PodCount: pods}
		mu.Unlock()

		return nil
	})

	return utils.WithoutNils(outputs), nil
}

func StatsWorker() error {
	clusterStats, err := GetClusterStats()
	if err != nil {
		return err
	}

	if clusterStats == nil {
		return wErrors.NoClustersErr
	}

	collected := body.SystemStats{K8sStats: body.K8sStats{PodCount: 0, Clusters: clusterStats}}
	for _, cluster := range clusterStats {
		collected.K8sStats.PodCount += cluster.PodCount
	}

	return system_stats_repo.New().Save(&body.TimestampedSystemStats{
		Stats:     collected,
		Timestamp: time.Now(),
	})
}

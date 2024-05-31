package system_state_poll

import (
	"go-deploy/models/model"
	errors2 "go-deploy/pkg/services/errors"
	"k8s.io/client-go/kubernetes"
	"log"
	"sync"
)

func ForEachHost(taskName string, hosts []model.Host, job func(worker int, host *model.Host) error) error {
	wg := sync.WaitGroup{}

	mutex := sync.RWMutex{}
	var failedHosts []string

	for idx, host := range hosts {
		wg.Add(1)

		i := idx
		h := host

		go func(i int) {
			err := job(i, &h)
			if err != nil {
				log.Printf("failed to execute task %s for host %s. details: %s\n", taskName, h.IP, err)
				mutex.Lock()
				failedHosts = append(failedHosts, h.Name)
				mutex.Unlock()
			}
			wg.Done()
		}(i)
	}

	wg.Wait()

	if len(failedHosts) > 0 {
		return errors2.NewFailedTaskErr(failedHosts)
	}

	return nil
}

func ForEachCluster(taskName string, clusters map[string]kubernetes.Clientset, job func(worker int, name string, cluster *kubernetes.Clientset) error) {
	wg := sync.WaitGroup{}

	idx := 0
	for name, cluster := range clusters {
		wg.Add(1)

		i := idx
		n := name
		c := cluster

		go func() {
			err := job(i, n, &c)
			if err != nil {
				log.Printf("failed to execute task %s for cluster %s. details: %s\n", taskName, n, err)
			}
			wg.Done()
		}()

		idx++
	}

	wg.Wait()
}

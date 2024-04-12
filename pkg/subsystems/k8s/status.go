package k8s

import (
	"context"
	"fmt"
	"go-deploy/pkg/subsystems/k8s/keys"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	v1 "kubevirt.io/api/core/v1"
	"time"
)

// SetupStatusWatcher sets up a status watcher for a given resource type in a namespace
//
// This should only be called once per cluster
func (client *Client) SetupStatusWatcher(ctx context.Context, resourceType string, handler func(string, string)) error {
	switch resourceType {
	case "vm":
		setupVmWatcher := func(namespace string) (watch.Interface, error) {
			statusChan, err := client.KubeVirtK8sClient.KubevirtV1().VirtualMachines(client.Namespace).Watch(ctx, metav1.ListOptions{Watch: true})
			if err != nil {
				return nil, err
			}

			if statusChan == nil {
				return nil, fmt.Errorf("failed to watch VirtualMachines, no channel returned")
			}

			return statusChan, nil
		}

		watcher, err := setupVmWatcher(client.Namespace)
		if err != nil {
			return err
		}

		resultsChan := watcher.ResultChan()
		go func() {
			// For now, the watch stops working sometimes, so we need to restart it every 10 seconds
			// This is a temporary fix until we find the root cause
			recreateInterval := time.Tick(20 * time.Second)

			for {
				select {
				case event := <-resultsChan:
					if event.Type == watch.Added || event.Type == watch.Modified {
						vm := event.Object.(*v1.VirtualMachine)

						// Fetch deploy name from label LabelDeployName
						deployName, ok := vm.Labels[keys.LabelDeployName]
						if ok {
							handler(deployName, string(vm.Status.PrintableStatus))
						}
					}
				case <-recreateInterval:
					watcher.Stop()
					watcher, err = setupVmWatcher(client.Namespace)
					if err != nil {
						fmt.Println("Failed to restart VM status watcher, sleeping for 10 seconds before retrying")
						time.Sleep(10 * time.Second)
						return
					}
					resultsChan = watcher.ResultChan()
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	return nil
}

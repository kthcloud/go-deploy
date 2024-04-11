package k8s

import (
	"context"
	"go-deploy/pkg/subsystems/k8s/keys"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	v1 "kubevirt.io/api/core/v1"
)

// SetupStatusWatcher sets up a status watcher for a given resource type in a namespace
//
// This should only be called once per cluster
func (client *Client) SetupStatusWatcher(ctx context.Context, resourceType string, handler func(string, string)) error {
	switch resourceType {
	case "vm":
		statusChan, err := client.KubeVirtK8sClient.KubevirtV1().VirtualMachines(client.Namespace).Watch(ctx, metav1.ListOptions{})

		if err != nil {
			return err
		}

		if statusChan == nil {
			return nil
		}

		resultsChan := statusChan.ResultChan()
		go func() {
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
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	return nil
}

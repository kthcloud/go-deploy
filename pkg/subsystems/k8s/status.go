package k8s

import (
	"context"
	"fmt"
	"go-deploy/pkg/subsystems/k8s/keys"
	"go-deploy/pkg/subsystems/k8s/models"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	v1 "kubevirt.io/api/core/v1"
	"time"
)

// SetupStatusWatcher sets up a status watcher for a given resource type in a namespace
//
// This should only be called once per cluster
func (client *Client) SetupStatusWatcher(ctx context.Context, resourceType string, handler func(string, interface{})) error {
	switch resourceType {
	case "deployment":
		return client.deploymentStatusWatcher(ctx, handler)
	case "vm":
		return client.vmStatusWatcher(ctx, handler)
	}

	return nil
}

func (client *Client) deploymentStatusWatcher(ctx context.Context, handler func(string, interface{})) error {
	setupDeploymentWatcher := func(namespace string) (watch.Interface, error) {
		statusChan, err := client.K8sClient.AppsV1().Deployments(client.Namespace).Watch(ctx, metav1.ListOptions{Watch: true})
		if err != nil {
			return nil, err
		}

		if statusChan == nil {
			return nil, fmt.Errorf("failed to watch deployments, no channel returned")
		}

		return statusChan, nil
	}

	watcher, err := setupDeploymentWatcher(client.Namespace)
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
					deployment := event.Object.(*appsv1.Deployment)

					// Fetch deploy name from label LabelDeployName
					deployName, ok := deployment.Labels[keys.LabelDeployName]
					if ok {
						var replicas int
						if deployment.Spec.Replicas != nil {
							replicas = int(*deployment.Spec.Replicas)
						} else {
							replicas = 1
						}

						handler(deployName, &models.DeploymentStatus{
							Generation:          int(deployment.Generation),
							DesiredReplicas:     replicas,
							ReadyReplicas:       int(deployment.Status.ReadyReplicas),
							AvailableReplicas:   int(deployment.Status.AvailableReplicas),
							UnavailableReplicas: int(deployment.Status.UnavailableReplicas),
						})
					}
				}
			case <-recreateInterval:
				watcher.Stop()
				watcher, err = setupDeploymentWatcher(client.Namespace)
				if err != nil {
					fmt.Println("Failed to restart Deployment status watcher, sleeping for 10 seconds before retrying")
					time.Sleep(10 * time.Second)
					return
				}
				resultsChan = watcher.ResultChan()
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (client *Client) vmStatusWatcher(ctx context.Context, handler func(string, interface{})) error {
	setupVmWatcher := func(namespace string) (watch.Interface, error) {
		statusChan, err := client.KubeVirtK8sClient.KubevirtV1().VirtualMachines(client.Namespace).Watch(ctx, metav1.ListOptions{Watch: true})
		if err != nil {
			return nil, err
		}

		if statusChan == nil {
			return nil, fmt.Errorf("failed to watch vms, no channel returned")
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
						handler(deployName, &models.VmStatus{
							PrintableStatus: string(vm.Status.PrintableStatus),
						})
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

	return nil
}

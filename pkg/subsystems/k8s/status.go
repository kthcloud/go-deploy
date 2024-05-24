package k8s

import (
	"context"
	"fmt"
	"go-deploy/pkg/log"
	"go-deploy/pkg/subsystems/k8s/keys"
	"go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/pkg/subsystems/k8s/opts"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	kubevirtv1 "kubevirt.io/api/core/v1"
	"regexp"
	"strings"
	"time"
)

// SetupStatusWatcher sets up a status watcher for a given resource type in a namespace
//
// This should only be called once per cluster
func (client *Client) SetupStatusWatcher(ctx context.Context, resourceType string, handler func(string, interface{}), opts ...opts.WatcherOpts) error {
	switch resourceType {
	case "deployment":
		return client.deploymentStatusWatcher(ctx, handler, opts...)
	case "vm":
		return client.vmStatusWatcher(ctx, handler, opts...)
	case "vmi":
		return client.vmiStatusWatcher(ctx, handler, opts...)
	case "event":
		return client.eventWatcher(ctx, handler, opts...)
	}

	return nil
}

func (client *Client) deploymentStatusWatcher(ctx context.Context, handler func(string, interface{}), opts ...opts.WatcherOpts) error {
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
					log.Println("Failed to restart Deployment status watcher, sleeping for 10 seconds before retrying")
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

func (client *Client) vmStatusWatcher(ctx context.Context, handler func(string, interface{}), opts ...opts.WatcherOpts) error {
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
					vm := event.Object.(*kubevirtv1.VirtualMachine)

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
					log.Println("Failed to restart VM status watcher, sleeping for 10 seconds before retrying")
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

func (client *Client) vmiStatusWatcher(ctx context.Context, handler func(string, interface{}), opts ...opts.WatcherOpts) error {
	setupVmWatcher := func(namespace string) (watch.Interface, error) {
		statusChan, err := client.KubeVirtK8sClient.KubevirtV1().VirtualMachineInstances(client.Namespace).Watch(ctx, metav1.ListOptions{Watch: true})
		if err != nil {
			return nil, err
		}

		if statusChan == nil {
			return nil, fmt.Errorf("failed to watch vmis, no channel returned")
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
					vmi := event.Object.(*kubevirtv1.VirtualMachineInstance)

					// Fetch deploy name from label LabelDeployName
					deployName, ok := vmi.Labels[keys.LabelDeployName]
					if ok {
						vmiStatus := &models.VmiStatus{}

						// Get the host name from the label "kubevirt.io/nodeName"
						host, ok := vmi.Labels["kubevirt.io/nodeName"]
						if ok {
							vmiStatus.Host = &host
						}

						handler(deployName, vmiStatus)
					}
				}
			case <-recreateInterval:
				watcher.Stop()
				watcher, err = setupVmWatcher(client.Namespace)
				if err != nil {
					log.Println("Failed to restart VM instance status watcher, sleeping for 10 seconds before retrying")
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

func (client *Client) eventWatcher(ctx context.Context, handler func(string, interface{}), opts ...opts.WatcherOpts) error {
	setupEventWatcher := func(namespace string) (watch.Interface, error) {
		statusChan, err := client.K8sClient.CoreV1().Events(client.Namespace).Watch(ctx, metav1.ListOptions{
			// Fields involvesObject.kind=Pod and type=Warning and reason=FailedMount or BackOff or Failed
			FieldSelector: "involvedObject.kind=Pod,type=Warning",
			Watch:         true,
		})
		if err != nil {
			return nil, err
		}

		if statusChan == nil {
			return nil, fmt.Errorf("failed to watch events, no channel returned")
		}

		return statusChan, nil
	}

	watcher, err := setupEventWatcher(client.Namespace)
	if err != nil {
		return err
	}

	resultsChan := watcher.ResultChan()
	go func() {
		// For now, the watch stops working sometimes, so we need to restart it every 300 seconds
		recreateInterval := time.Tick(300 * time.Second)

		for {
			select {
			case event := <-resultsChan:
				if event.Type == watch.Added || event.Type == watch.Modified {
					e := event.Object.(*corev1.Event)

					go func(e *corev1.Event) {
						var deployName string
						var objectKind string
						var reason string
						var description string

						switch e.InvolvedObject.Kind {
						case "Pod":
							objectKind = models.EventObjectKindDeployment
							// InvolvedObject.Name is a pod name, we need to get the deployment name
							pod, err := client.K8sClient.CoreV1().Pods(client.Namespace).Get(ctx, e.InvolvedObject.Name, metav1.GetOptions{})
							if err != nil {
								return
							}
							// Fetch deploy name from label LabelDeployName
							var ok bool
							deployName, ok = pod.Labels[keys.LabelDeployName]
							if !ok {
								return
							}
						}
						if objectKind == "" {
							return
						}

						switch e.Reason {
						case "FailedMount":
							reason = models.EventReasonMountFailed
							// We parse out a nicer message for mount failures
							// Extract the following substring "MountVolume.SetUp failed for volume "<anyname>" : mount failed"
							// from the message using regex
							var exitStatus string
							exitStatusRegex := regexp.MustCompile(`exit status (\d+)`)
							exitStatusMatches := exitStatusRegex.FindStringSubmatch(e.Message)
							if len(exitStatusMatches) > 1 {
								exitStatus = fmt.Sprintf(" (Exit code: %s)", exitStatusMatches[1])
							}

							var volumeName string
							regex := regexp.MustCompile(`MountVolume\.SetUp failed for volume "(.*)" : mount failed`)
							matches := regex.FindStringSubmatch(e.Message)
							if len(matches) > 1 {
								// If the volume name starts with deployName, we remove it, local-test-my-volume => my-volume
								volumeName = fmt.Sprintf(" %s", strings.TrimPrefix(matches[1], deployName+"-"))
							}
							description = fmt.Sprintf("Failed to mount volume%s. Ensure the path exists in your storage%s.", volumeName, exitStatus)

						case "BackOff":
							reason = models.EventReasonCrashLoop
							// We parse out a nicer message for crash loops
							// BackOff can occur in many cases, but we handle those in other branches
							if !strings.Contains(e.Message, "Back-off restarting failed container") {
								return
							}

							description = "Crash loop detected. Ensure your deployment is configured correctly."
						case "Failed":
							// Failed can be due to image pull failures or image pull backoff
							// If not, we ignore it
							if !strings.Contains(e.Message, "Failed to pull image") {
								return
							}

							reason = models.EventReasonImagePullFailed
							// We parse out a nicer message for image pull failures
							// Extract the requested image from "Failed to pull image "thisdoesnotexist"" using regex
							regex := regexp.MustCompile(`Failed to pull image "(.*)": failed to pull and unpack image`)
							matches := regex.FindStringSubmatch(e.Message)
							var image string
							if len(matches) > 1 {
								image = fmt.Sprintf(" %s", matches[1])
							}

							description = fmt.Sprintf("Failed to pull image%s. Ensure the image exists and is accessible.", image)
						default:
							return
						}

						var eventType string
						switch e.Type {
						case "Warning":
							eventType = models.EventTypeWarning
						case "Normal":
							eventType = models.EventTypeNormal
						}

						// Fetch deploy name from label LabelDeployName
						handler(deployName, &models.Event{
							Type:        eventType,
							Reason:      reason,
							Description: description,
							ObjectKind:  objectKind,
						})
					}(e)
				}
			case <-recreateInterval:
				watcher.Stop()
				watcher, err = setupEventWatcher(client.Namespace)
				if err != nil {
					log.Println("Failed to restart Event status watcher, sleeping for 10 seconds before retrying")
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

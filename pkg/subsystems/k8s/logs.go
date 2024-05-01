package k8s

import (
	"bufio"
	"context"
	"fmt"
	"go-deploy/pkg/log"
	"go-deploy/pkg/subsystems/k8s/keys"
	"go-deploy/utils"
	"golang.org/x/exp/maps"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"sort"
	"strings"
	"time"
)

const (
	// PodEventStart is emitted when a pod starts
	PodEventStart = "start"
	// PodEventStop is emitted when a pod stops
	PodEventStop = "stop"
)

type PodEventType struct {
	deploymentName string
	podName        string
	event          string
	startTime      time.Time
}

func PodEvent(deploymentName, podName, event string, startTime time.Time) PodEventType {
	return PodEventType{deploymentName: deploymentName, podName: podName, event: event, startTime: startTime}
}

// getPodNames gets the names of all pods for a deployment
// This is used when setting up a log stream for a deployment
func (client *Client) getPodNames(namespace, deploymentName string) ([]string, error) {
	pods, err := client.K8sClient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", keys.LabelDeployName, deploymentName),
	})
	if err != nil {
		return nil, err
	}

	podNames := make([]string, 0)
	for _, pod := range pods.Items {
		if pod.Status.Phase != v1.PodRunning {
			continue
		}
		podNames = append(podNames, pod.Name)
	}

	return podNames, nil
}

// SetupLogStream sets up a log stream for the entire namespace
//
// This should only be called once per cluster
func (client *Client) SetupLogStream(ctx context.Context, allowedNames []string, handler func(string, string, int, time.Time)) error {
	_ = func(err error) error {
		return fmt.Errorf("failed to create k8s log stream. details: %w", err)
	}

	go func() {
		// activeStreams is a map of active log streams structured as map[deploymentName]map[podName]podNumber
		activeStreams := make(map[string]map[string]int)
		cancelFuncs := make(map[string]map[string]context.CancelFunc)
		podChannel := make(chan PodEventType, 100)

		// Convert allowedNames to a map for faster lookups
		allowedNamesMap := make(map[string]struct{})
		for _, name := range allowedNames {
			allowedNamesMap[name] = struct{}{}
		}

		factory := informers.NewSharedInformerFactoryWithOptions(client.K8sClient, 0, informers.WithNamespace(client.Namespace))
		podInformer := factory.Core().V1().Pods().Informer()

		// Returns the name of the deployment, when it was created and whether the pod is allowed
		allowedPod := func(pod *v1.Pod) (string, time.Time, bool) {
			var deploymentName string
			var ok bool

			if deploymentName, ok = pod.Labels[keys.LabelDeployName]; !ok {
				return "", time.Time{}, false
			}

			if _, ok = allowedNamesMap[deploymentName]; !ok {
				return "", time.Time{}, false
			}

			allowedStatuses := []v1.PodPhase{
				v1.PodRunning,
				v1.PodFailed,
			}

			allowed := false
			for _, status := range allowedStatuses {
				if pod.Status.Phase == status {
					allowed = true
					break
				}
			}

			if !allowed {
				return "", time.Time{}, false
			}

			return deploymentName, pod.CreationTimestamp.Time, true
		}

		_, err := podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				pod, ok := obj.(*v1.Pod)
				if !ok {
					return
				}

				deploymentName, createdAt, ok := allowedPod(pod)
				if !ok {
					return
				}

				podChannel <- PodEvent(deploymentName, pod.Name, PodEventStart, createdAt)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				pod, ok := newObj.(*v1.Pod)
				if !ok {
					return
				}

				deploymentName, createdAt, ok := allowedPod(pod)
				if !ok {
					return
				}

				podChannel <- PodEvent(deploymentName, pod.Name, PodEventStart, createdAt)
			},
			DeleteFunc: func(obj interface{}) {
				pod, ok := obj.(*v1.Pod)
				if !ok {
					return
				}

				var deploymentName string
				if deploymentName, ok = pod.Labels[keys.LabelDeployName]; !ok {
					return
				}

				if _, ok = allowedNamesMap[deploymentName]; !ok {
					return
				}

				podChannel <- PodEvent(deploymentName, pod.Name, PodEventStop, time.Time{})
			},
		})
		if err != nil {
			return
		}

		factory.Start(ctx.Done())
		factory.WaitForCacheSync(ctx.Done())

		for {
			select {
			case <-ctx.Done():
				for _, cancelFuncMap := range cancelFuncs {
					for _, cancelFunc := range cancelFuncMap {
						cancelFunc()
					}
				}
				return
			case e := <-podChannel:
				switch e.event {
				case PodEventStart:
					// Create deployment map if it does not exist
					if _, ok := activeStreams[e.deploymentName]; !ok {
						activeStreams[e.deploymentName] = make(map[string]int)
					}

					// Create cancel func map if it does not exist
					if _, ok := cancelFuncs[e.deploymentName]; !ok {
						cancelFuncs[e.deploymentName] = make(map[string]context.CancelFunc)
					}

					// Check if pod is already being streamed
					if _, ok := activeStreams[e.deploymentName][e.podName]; ok {
						continue
					}

					// Add pod to deployment map
					idx := getFreePodNumber(activeStreams[e.deploymentName])
					activeStreams[e.deploymentName][e.podName] = idx

					cancelCtx, cancelFunc := context.WithCancel(ctx)
					cancelFuncs[e.deploymentName][e.podName] = cancelFunc

					log.Println("Starting logger for pod", e.podName, "with idx", idx)

					go func() {
						client.readLogs(cancelCtx, idx, client.Namespace, e.deploymentName, e.podName, e.startTime, podChannel, handler)
					}()

				case PodEventStop:
					// Stop the log stream for the pod
					cancelFunc, ok := cancelFuncs[e.deploymentName][e.podName]
					if ok {
						log.Println("Stopping logger for", e.podName)

						cancelFunc()
						delete(cancelFuncs[e.deploymentName], e.podName)

						delete(activeStreams[e.deploymentName], e.podName)
					}
				}
			}
		}
	}()

	return nil
}

// readLogs reads logs from a pod and sends them to the handler
// It listens to the PodEventType channel to know when to stop, and emits a PodEventStop event when it stops
func (client *Client) readLogs(ctx context.Context, podNumber int, namespace, deploymentName, podName string, start time.Time, eventChan chan PodEventType, handler func(string, string, int, time.Time)) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				utils.PrettyPrintError(fmt.Errorf("failed to read logs for pod %s (err). details: %w", podName, err))
			} else {
				utils.PrettyPrintError(fmt.Errorf("failed to read logs for pod %s (panic). details: %v", podName, r))
			}
			eventChan <- PodEventType{event: PodEventStop, podName: podName}
		}
	}()

	logStream, err := getK8sLogStream(client, namespace, podName, start)
	if err != nil {
		if IsNotFoundErr(err) {
			// Pod got deleted for some reason, so we just stop the log stream
			return
		}

		utils.PrettyPrintError(fmt.Errorf("failed to create k8s log stream for pod %s. details: %w", podName, err))
		return
	}
	defer func(logStream io.ReadCloser) {
		if logStream != nil {
			_ = logStream.Close()
		}
	}(logStream)

	reader := bufio.NewScanner(logStream)

	var line string
	for {
		select {
		case <-ctx.Done():
			return
		default:
			for reader.Scan() {
				if ctx.Err() != nil {
					break
				}

				line = reader.Text()
				if isExitLine(line) {
					break
				}

				handler(line, deploymentName, podNumber, time.Now())
			}

			time.Sleep(100 * time.Millisecond)
		}
	}
}

// getK8sLogStream gets a log stream for a pod in Kubernetes
func getK8sLogStream(client *Client, namespace, podName string, since time.Time) (io.ReadCloser, error) {
	var t *metav1.Time
	if !since.IsZero() {
		t = &metav1.Time{Time: since}
	}

	podLogsConnection := client.K8sClient.CoreV1().Pods(namespace).GetLogs(podName, &v1.PodLogOptions{
		Follow:    true,
		SinceTime: t,
	})

	logStream, err := podLogsConnection.Stream(context.Background())
	if err != nil {
		return nil, err
	}

	return logStream, nil
}

// isExitLine is a helper function to determine if a line from a log stream is an exit line.
// This is used to know when to stop reading logs from a pod, since Kubernetes does not provide a way
// to do this directly (at least in a reasonable amount of time).
func isExitLine(line string) bool {
	firstPart := strings.Contains(line, "rpc error: code = NotFound desc = an error occurred when try to find container")
	lastPart := strings.Contains(line, "not found")

	notYetStarted := firstPart && lastPart

	sigQuit := strings.HasSuffix(line, "signal 3 (SIGQUIT) received, shutting down")
	gracefullyShuttingDown := strings.HasSuffix(line, ": gracefully shutting down")

	return notYetStarted || sigQuit || gracefullyShuttingDown
}

// getFreePodNumber is a helper function that gets the next free pod number for a deployment
// The number is used as a unique nice-to-read identifier for a pod.
func getFreePodNumber(activeStreams map[string]int) int {
	values := maps.Values(activeStreams)

	sort.Slice(values, func(i, j int) bool {
		return values[i] < values[j]
	})

	for i := 0; i < len(values); i++ {
		if i != values[i] {
			return i
		}
	}

	return len(values)
}

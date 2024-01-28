package k8s

import (
	"bufio"
	"context"
	"fmt"
	"go-deploy/pkg/subsystems/k8s/keys"
	"go-deploy/utils"
	"golang.org/x/exp/maps"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"log"
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

type podEvent struct {
	podName string
	event   string
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

// SetupLogStream sets up a log stream for the entire cluster
//
// This should only be called once per cluster
func (client *Client) SetupLogStream(ctx context.Context, deploymentName string, handler func(string, int, time.Time)) error {
	_ = func(err error) error {
		return fmt.Errorf("failed to create k8s log stream. details: %w", err)
	}

	go func() {
		activeStreams := make(map[string]int)
		cancelFuncs := make(map[string]context.CancelFunc)
		podChannel := make(chan podEvent, 100)

		factory := informers.NewSharedInformerFactoryWithOptions(client.K8sClient, 0, informers.WithNamespace(client.Namespace))
		podInformer := factory.Core().V1().Pods().Informer()

		podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				pod := obj.(*v1.Pod)

				if label, ok := pod.Labels[keys.LabelDeployName]; !ok || label != deploymentName {
					return
				}

				if pod.Status.Phase != v1.PodRunning {
					return
				}

				podChannel <- podEvent{event: PodEventStart, podName: pod.Name}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				pod := newObj.(*v1.Pod)
				if label, ok := pod.Labels[keys.LabelDeployName]; !ok || label != deploymentName {
					return
				}

				if pod.Status.Phase != v1.PodRunning {
					return
				}

				podChannel <- podEvent{event: PodEventStart, podName: pod.Name}
			},
			DeleteFunc: func(obj interface{}) {
				pod := obj.(*v1.Pod)
				podChannel <- podEvent{event: PodEventStop, podName: pod.Name}
			},
		})

		factory.Start(ctx.Done())
		factory.WaitForCacheSync(ctx.Done())

		for {
			select {
			case <-ctx.Done():
				for _, cancelFunc := range cancelFuncs {
					cancelFunc()
				}
				return
			case e := <-podChannel:
				switch e.event {
				case PodEventStart:
					if _, ok := activeStreams[e.podName]; ok {
						continue
					}

					idx := getFreePodNumber(activeStreams)
					activeStreams[e.podName] = idx

					cancelCtx, cancelFunc := context.WithCancel(ctx)
					cancelFuncs[e.podName] = cancelFunc

					log.Println("starting logger for pod", e.podName, "with idx", idx)

					go func() {
						client.readLogs(cancelCtx, idx, client.Namespace, e.podName, podChannel, handler)
					}()

				case PodEventStop:
					cancelFunc, ok := cancelFuncs[e.podName]
					if ok {
						log.Println("stopping logger for pod", e.podName)

						cancelFunc()
						delete(cancelFuncs, e.podName)

						delete(activeStreams, e.podName)
					}
				}
			}
		}
	}()

	return nil
}

// readLogs reads logs from a pod and sends them to the handler
// It listens to the podEvent channel to know when to stop, and emits a PodEventStop event when it stops
func (client *Client) readLogs(ctx context.Context, podNumber int, namespace, podName string, eventChan chan podEvent, handler func(string, int, time.Time)) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				utils.PrettyPrintError(fmt.Errorf("failed to read logs for pod %s (err). details: %w", podName, err))
			} else {
				utils.PrettyPrintError(fmt.Errorf("failed to read logs for pod %s (panic). details: %v", podName, r))
			}
			eventChan <- podEvent{event: PodEventStop, podName: podName}
		}
	}()

	logStream, err := getK8sLogStream(client, namespace, podName, 0)
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

				handler(line, podNumber, time.Now())
			}

			time.Sleep(100 * time.Millisecond)
		}
	}
}

// getK8sLogStream gets a log stream for a pod in Kubernetes
func getK8sLogStream(client *Client, namespace, podName string, history int) (io.ReadCloser, error) {
	podLogsConnection := client.K8sClient.CoreV1().Pods(namespace).GetLogs(podName, &v1.PodLogOptions{
		Follow:    true,
		TailLines: &[]int64{int64(history)}[0],
	})

	logStream, err := podLogsConnection.Stream(context.Background())
	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("failed to create k8s log stream for pod %s. details: %w", podName, err))
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

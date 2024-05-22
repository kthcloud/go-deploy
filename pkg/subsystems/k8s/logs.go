package k8s

import (
	"bufio"
	"context"
	"fmt"
	"go-deploy/pkg/log"
	"go-deploy/pkg/subsystems/k8s/keys"
	"go-deploy/pkg/subsystems/k8s/models"
	"golang.org/x/exp/maps"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
	"strings"
	"time"
)

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

// SetupLogStream reads logs from a pod and sends them to the callback function
func (client *Client) SetupLogStream(ctx context.Context, podName string, from time.Time, onLog func(deploymentName string, lines []models.LogLine)) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to set up log stream for pod %s. details: %w", podName, err)
	}

	deploymentName := client.getDeploymentName(podName)
	if deploymentName == "" {
		return makeError(fmt.Errorf("deployment name not found for pod %s", podName))
	}

	logStream, err := client.getPodLogStream(ctx, client.Namespace, podName, from)
	if err != nil {
		if IsNotFoundErr(err) {
			// Pod got deleted for some reason, so we just stop the log stream
			return nil
		}

		return makeError(err)
	}

	go func() {
		defer log.Println("Log stream for pod", podName, "stopped")

		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		defer cancel()

		defer func(logStream io.ReadCloser) {
			if logStream != nil {
				_ = logStream.Close()
			}
		}(logStream)

		reader := bufio.NewScanner(logStream)

		lines := make([]models.LogLine, 0, 10)
		lastPush := time.Now()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				for reader.Scan() {
					if ctx.Err() != nil {
						return
					}

					line := reader.Text()
					if isExitLine(line) {
						if len(lines) > 0 {
							onLog(deploymentName, lines)
							lines = nil
						}

						return
					}

					lines = append(lines, models.LogLine{
						Line:      line,
						CreatedAt: time.Now(),
					})

					// Push logs every 5 seconds or when the buffer is full (10 lines)
					if time.Since(lastPush) > 5*time.Second || len(lines) >= 10 {
						onLog(deploymentName, lines)
						lines = nil
						lastPush = time.Now()
					}
				}
			}
		}
	}()

	return nil
}

// getDeploymentName gets the name of a deployment from a pod
func (client *Client) getDeploymentName(podName string) string {
	pod, err := client.K8sClient.CoreV1().Pods(client.Namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return ""
	}

	deploymentName, ok := pod.Labels[keys.LabelDeployName]
	if !ok {
		return ""
	}

	return deploymentName
}

// getPodLogStream gets a log stream for a pod in Kubernetes
func (client *Client) getPodLogStream(ctx context.Context, namespace, podName string, since time.Time) (io.ReadCloser, error) {
	var t *metav1.Time
	if !since.IsZero() {
		t = &metav1.Time{Time: since}
	}

	podLogsConnection := client.K8sClient.CoreV1().Pods(namespace).GetLogs(podName, &v1.PodLogOptions{
		Follow:    true,
		SinceTime: t,
	})

	logStream, err := podLogsConnection.Stream(ctx)
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

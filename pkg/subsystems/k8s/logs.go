package k8s

import (
	"bufio"
	"context"
	"fmt"
	"go-deploy/pkg/subsystems/k8s/keys"
	"go-deploy/utils"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"strings"
	"sync"
	"time"
)

func (client *Client) getPodNames(namespace, deploymentID string) ([]string, error) {
	pods, err := client.K8sClient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", keys.ManifestLabelID, deploymentID),
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

func (client *Client) setupPodLogStreamer(ctx context.Context, namespace, deploymentID string, handler func(string, int, time.Time)) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s log stream for deployment %s. details: %w", deploymentID, err)
	}

	activeStreams := make(map[string]int)
	mutex := sync.Mutex{}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(500 * time.Millisecond)

			podNames, err := client.getPodNames(namespace, deploymentID)
			if err != nil {
				utils.PrettyPrintError(makeError(err))
				return
			}

			if len(podNames) == 0 {
				return
			}

			// setup log stream for each pod
			for idx, podName := range podNames {
				if _, ok := activeStreams[podName]; ok {
					continue
				}

				mutex.Lock()
				activeStreams[podName] = getFreePodNumber(activeStreams)
				mutex.Unlock()

				podNumber := idx
				localPodName := podName
				go func() {
					client.readLogs(ctx, podNumber, namespace, localPodName, handler)
				}()
			}

			// remove log stream for each pod that is not running
			for podName := range activeStreams {
				found := false
				for _, name := range podNames {
					if name == podName {
						found = true
						break
					}
				}

				if !found {
					mutex.Lock()
					delete(activeStreams, podName)
					mutex.Unlock()
				}
			}
		}
	}
}

func (client *Client) readLogs(ctx context.Context, podNumber int, namespace, podName string, handler func(string, int, time.Time)) {
	var logStream io.ReadCloser
	defer func(logStream io.ReadCloser) {
		if logStream != nil {
			_ = logStream.Close()
		}
	}(logStream)

	createNewStream := true
	var reader *bufio.Scanner

	var line string
	for {
		select {
		case <-ctx.Done():
			log.Println("logger stopped for pod", podName)
			return
		default:
			if createNewStream {
				logStream, err := getK8sLogStream(client, namespace, podName, 0)
				if err != nil {
					if IsNotFoundErr(err) {
						// pod got deleted for some reason, so we just stop the log stream
						return
					}

					utils.PrettyPrintError(fmt.Errorf("failed to create k8s log stream for pod %s. details: %w", podName, err))
					return
				}
				reader = bufio.NewScanner(logStream)

				createNewStream = false
			}

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

func isExitLine(line string) bool {
	firstPart := strings.Contains(line, "rpc error: code = NotFound desc = an error occurred when try to find container")
	lastPart := strings.Contains(line, "not found")

	notYetStarted := firstPart && lastPart

	sigQuit := strings.HasSuffix(line, "signal 3 (SIGQUIT) received, shutting down")
	gracefullyShuttingDown := strings.HasSuffix(line, ": gracefully shutting down")

	return notYetStarted || sigQuit || gracefullyShuttingDown
}

func (client *Client) SetupDeploymentLogStream(ctx context.Context, deploymentID string, handler func(string, int, time.Time)) error {
	go client.setupPodLogStreamer(ctx, client.Namespace, deploymentID, handler)
	return nil
}

func getFreePodNumber(activeStreams map[string]int) int {
	max := 0
	for _, v := range activeStreams {
		if v > max {
			max = v
		}
	}

	return max + 1
}

// 2023/12/03 19:11:23 [notice] 1#1: signal 3 (SIGQUIT) received, shutting down
// 2023/12/03 19:11:23 [notice] 21#21: gracefully shutting down
// 2023/12/03 19:11:23 [notice] 20#20: exiting
// 2023/12/03 19:11:23 [notice] 22#22: exit

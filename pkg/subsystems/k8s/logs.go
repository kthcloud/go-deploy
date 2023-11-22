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

	podNames := make([]string, len(pods.Items))
	for idx, pod := range pods.Items {
		podNames[idx] = pod.Name
	}

	return podNames, nil
}

func (client *Client) setupPodLogStreamer(ctx context.Context, namespace, deploymentID string, handler func(string, string)) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s log stream for deployment %s. details: %w", deploymentID, err)
	}

	podNames, err := client.getPodNames(namespace, deploymentID)
	if err != nil {
		utils.PrettyPrintError(makeError(err))
		return
	}

	if len(podNames) == 0 {
		return
	}

	wg := sync.WaitGroup{}
	for idx, podName := range podNames {
		wg.Add(1)

		localIdx := idx
		localPodName := podName
		go func() {
			client.readLogs(ctx, fmt.Sprintf("[pod %d]", localIdx), namespace, localPodName, handler)
			wg.Done()
		}()
	}
	wg.Wait()

	return
}

func (client *Client) readLogs(ctx context.Context, prefix, namespace, podName string, handler func(string, string)) {
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
				logStream, err := getK8sLogStream(client, namespace, podName, 20)
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
					createNewStream = true
					break
				}

				handler(prefix, line)
			}

			time.Sleep(100 * time.Millisecond)
			handler("[control]", "__control")
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
	return firstPart && lastPart
}

func (client *Client) SetupLogStream(ctx context.Context, namespace, deploymentID string, handler func(string, string)) error {
	go client.setupPodLogStreamer(ctx, namespace, deploymentID, handler)
	return nil
}

package k8s

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"strings"
)

func (client *Client) getPodName(namespace, deployment string) (string, error) {
	pods, err := client.K8sClient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", "app.kubernetes.io/name", deployment),
	})
	if err != nil {
		return "", err
	}

	// Only one pod is allowed
	for _, pod := range pods.Items {
		return pod.Name, nil
	}
	return "", errors.New("no pods in namespace")
}

func (client *Client) setupPodLogStreamer(cancelCtx context.Context, namespace, deployment string, handler func(string)) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s log stream for deployment %s. details: %s", deployment, err)
	}

	for {
		podName, err := client.getPodName(namespace, deployment)
		if err != nil {
			log.Println(makeError(err))
			return
		}

		if podName == "" {
			return
		}

		finished := client.readLogs(cancelCtx, namespace, podName, handler)
		if finished {
			return
		}
	}
}

func (client *Client) readLogs(cancelCtx context.Context, namespace, podName string, handler func(string)) bool {
	podLogsConnection := client.K8sClient.CoreV1().Pods(namespace).GetLogs(podName, &v1.PodLogOptions{
		Follow:    true,
		TailLines: &[]int64{int64(10)}[0],
	})
	logStream, err := podLogsConnection.Stream(context.Background())
	if err != nil {
		log.Println(fmt.Errorf("failed to create k8s log stream for pod %s. details: %s", podName, err))
		return true
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
		case <-cancelCtx.Done():
			return true
		default:
			for reader.Scan() {
				line = reader.Text()
				if isExitLine(line) {
					return false
				} else {
					handler(line)
				}
			}
		}
	}
}

func isExitLine(line string) bool {
	firstPart := strings.Contains(line, "rpc error: code = NotFound desc = an error occurred when try to find container")
	lastPart := strings.Contains(line, "not found")
	return firstPart && lastPart
}

func (client *Client) SetupLogStream(context context.Context, namespace, deployment string, handler func(string)) error {
	go client.setupPodLogStreamer(context, namespace, deployment, handler)
	return nil
}

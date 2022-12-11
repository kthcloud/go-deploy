package k8s

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (client *Client) getPodName(namespace string) (*string, error) {
	pods, err := client.K8sClient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Only one pod is allowed
	for _, pod := range pods.Items {
		return &pod.Name, nil
	}
	return nil, errors.New("no pods in namespace")
}

func (client *Client) getPodLogs(cancelCtx context.Context, namespace, podName string, handler func(string)) {
	podLogsConnection := client.K8sClient.CoreV1().Pods(namespace).GetLogs(podName, &v1.PodLogOptions{
		Follow:    true,
		TailLines: &[]int64{int64(10)}[0],
	})
	logStream, _ := podLogsConnection.Stream(context.Background())
	defer func(logStream io.ReadCloser) {
		_ = logStream.Close()
	}(logStream)

	reader := bufio.NewScanner(logStream)
	var line string
	for {
		select {
		case <-cancelCtx.Done():
			break
		default:
			for reader.Scan() {
				line = reader.Text()
				handler(line)
			}
		}
	}
}

func (client *Client) GetLogStream(context context.Context, namespace, name string, handler func(string)) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s log stream for deployment %s. details: %s", name, err)
	}

	podName, err := client.getPodName(namespace)
	if err != nil {
		return makeError(err)
	}

	if podName == nil {
		return makeError(fmt.Errorf("failed to find pod name for %s", name))
	}

	go client.getPodLogs(context, namespace, *podName, handler)

	return nil
}

package k8s

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"go-deploy/utils/subsystemutils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getPodName(name string) (*string, error) {
	prefixedName := subsystemutils.GetPrefixedName(name)
	pods, err := client.CoreV1().Pods(prefixedName).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Only one pod is allowed
	for _, pod := range pods.Items {
		return &pod.Name, nil
	}
	return nil, errors.New("no pods in namespace")
}

func getPodLogs(cancelCtx context.Context, name, podName string, handler func(string)) {
	podLogsConnection := client.CoreV1().Pods(subsystemutils.GetPrefixedName(name)).GetLogs(podName, &v1.PodLogOptions{
		Follow:    true,
		TailLines: &[]int64{int64(10)}[0],
	})
	logStream, _ := podLogsConnection.Stream(context.Background())
	defer logStream.Close()

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

func GetLogStream(context context.Context, name string, handler func(string)) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s log stream for deployment %s. details: %s", name, err)
	}

	podName, err := getPodName(name)
	if err != nil {
		return makeError(err)
	}

	if podName == nil {
		return makeError(fmt.Errorf("failed to find pod name for %s", name))
	}

	go getPodLogs(context, name, *podName, handler)

	return nil
}

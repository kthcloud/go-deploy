package k8s

import (
	"context"
	"fmt"
	"go-deploy/pkg/subsystems/k8s/manifests"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (client *Client) ServiceCreated(namespace, name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if k8s service %s is created. details: %s", name, err)
	}

	service, err := client.K8sClient.CoreV1().Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if isNotFoundError(err) {
			return false, nil
		}
		return false, makeError(err)
	}
	return service != nil, err
}

func (client *Client) CreateService(namespace, name string, port, targetPort int) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to k8s service %s. details: %s", name, err)
	}

	serviceCreated, err := client.ServiceCreated(namespace, name)
	if err != nil {
		return makeError(err)
	}

	if serviceCreated {
		return nil
	}

	namespaceCreated, err := client.NamespaceCreated(namespace)
	if err != nil {
		return makeError(err)
	}

	if !namespaceCreated {
		err = fmt.Errorf("no namespace created")
		return makeError(err)
	}

	service := manifests.CreateServiceManifest(namespace, name, port, targetPort)

	_, err = client.K8sClient.CoreV1().Services(namespace).Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		return makeError(err)
	}

	return nil
}

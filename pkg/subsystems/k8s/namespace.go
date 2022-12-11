package k8s

import (
	"context"
	"fmt"
	"go-deploy/pkg/subsystems/k8s/manifests"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (client *Client) NamespaceCreated(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if k8s namespace %s is created. details: %s", name, err)
	}

	namespace, err := client.K8sClient.CoreV1().Namespaces().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if isNotFoundError(err) {
			return false, nil
		}
		return false, makeError(err)
	}
	return namespace != nil, err
}

func (client *Client) NamespaceDeleted(name string) (bool, error) {
	created, err := client.NamespaceCreated(name)
	return !created, err
}

func (client *Client) CreateNamespace(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s namespace %s. details: %s", name, err)
	}

	namespaceCreated, err := client.NamespaceCreated(name)
	if err != nil {
		return makeError(err)
	}

	if namespaceCreated {
		return nil
	}

	namespace := manifests.CreateNamespaceManifest(name)

	_, err = client.K8sClient.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
	if err != nil {
		return makeError(err)
	}

	return err
}

func (client *Client) DeleteNamespace(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s namespace %s. details: %s", name, err)
	}

	namespaceDelted, err := client.NamespaceDeleted(name)
	if err != nil {
		return makeError(err)
	}

	if namespaceDelted {
		return nil
	}

	err = client.K8sClient.CoreV1().Namespaces().Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return makeError(err)
	}

	return nil
}

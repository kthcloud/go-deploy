package k8s

import (
	"context"
	"fmt"
	"go-deploy/pkg/subsystems/k8s/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func (client *Client) createNamespaceWatcher(ctx context.Context, resourceName string) (watch.Interface, error) {
	labelSelector := fmt.Sprintf("%s=%s", manifestLabelName, resourceName)

	opts := metav1.ListOptions{
		TypeMeta:      metav1.TypeMeta{},
		LabelSelector: labelSelector,
		FieldSelector: "",
	}

	return client.K8sClient.CoreV1().Namespaces().Watch(ctx, opts)
}

func (client *Client) waitNamespaceDeleted(ctx context.Context, resourceName string) error {
	watcher, err := client.createNamespaceWatcher(ctx, resourceName)
	if err != nil {
		return err
	}

	defer watcher.Stop()

	for {
		select {
		case event := <-watcher.ResultChan():

			if event.Type == watch.Deleted {
				return nil
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func (client *Client) NamespaceCreated(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if namespace %s is created. details: %s", name, err)
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

func (client *Client) CreateNamespace(public *models.NamespacePublic) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create namespace %s. details: %s", public.Name, err)
	}

	namespaceCreated, err := client.NamespaceCreated(public.Name)
	if err != nil {
		return makeError(err)
	}

	if namespaceCreated {
		return nil
	}

	namespace := CreateNamespaceManifest(public)

	_, err = client.K8sClient.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
	if err != nil {
		return makeError(err)
	}

	return err
}

func (client *Client) DeleteNamespace(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete namespace %s. details: %s", name, err)
	}

	namespaceDeleted, err := client.NamespaceDeleted(name)
	if err != nil {
		return makeError(err)
	}

	if namespaceDeleted {
		return nil
	}

	err = client.K8sClient.CoreV1().Namespaces().Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return makeError(err)
	}

	err = client.waitNamespaceDeleted(context.TODO(), name)
	if err != nil {
		return makeError(err)
	}

	return nil
}

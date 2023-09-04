package k8s

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/pkg/subsystems/k8s/keys"
	"go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/utils/subsystemutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func (client *Client) createNamespaceWatcher(ctx context.Context, resourceName string) (watch.Interface, error) {
	labelSelector := fmt.Sprintf("%s=%s", "kubernetes.io/metadata.name", resourceName)

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
	_ = func(err error) error {
		return fmt.Errorf("failed to check if namespace %s is created. details: %w", name, err)
	}

	if name == "" {
		return false, nil
	}

	list, err := client.K8sClient.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return false, err
	}

	for _, item := range list.Items {
		if item.Name == name {
			return true, nil
		}
	}

	return false, nil
}

func (client *Client) NamespaceDeleted(name string) (bool, error) {
	created, err := client.NamespaceCreated(name)
	return !created, err
}

func (client *Client) ReadNamespace(id string) (*models.NamespacePublic, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to read namespace %s. details: %w", id, err)
	}

	if id == "" {
		return nil, nil
	}

	list, err := client.K8sClient.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, item := range list.Items {
		idLabel := GetLabel(item.ObjectMeta.Labels, keys.ManifestLabelID)
		if idLabel == id {
			return models.CreateNamespacePublicFromRead(&item), nil
		}
	}

	return nil, nil
}

func (client *Client) CreateNamespace(public *models.NamespacePublic) (string, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to create namespace %s. details: %w", public.Name, err)
	}

	if public.Name == "" {
		return "", nil
	}

	// find if namespace already exists by name label
	list, err := client.K8sClient.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	for _, item := range list.Items {
		if FindLabel(item.ObjectMeta.Labels, keys.ManifestLabelName, public.Name) {
			idLabel := GetLabel(item.ObjectMeta.Labels, keys.ManifestLabelID)
			if idLabel != "" {
				return idLabel, nil
			}
		}
	}

	public.ID = uuid.New().String()
	public.FullName = subsystemutils.GetPrefixedName(public.Name)

	manifest := CreateNamespaceManifest(public)
	_, err = client.K8sClient.CoreV1().Namespaces().Create(context.TODO(), manifest, metav1.CreateOptions{})
	if err != nil {
		return "", err
	}

	return public.ID, nil
}

func (client *Client) DeleteNamespace(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete namespace %s. details: %w", name, err)
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

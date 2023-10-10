package k8s

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/pkg/subsystems/k8s/keys"
	"go-deploy/pkg/subsystems/k8s/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"log"
	"time"
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

func (client *Client) CreateNamespace(public *models.NamespacePublic) (*models.NamespacePublic, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to create namespace %s. details: %w", public.Name, err)
	}

	ns, err := client.K8sClient.CoreV1().Namespaces().Get(context.TODO(), public.Name, metav1.GetOptions{})
	if err == nil && !IsNotFoundErr(err) {
		return models.CreateNamespacePublicFromRead(ns), nil
	}

	if ns != nil {
		return models.CreateNamespacePublicFromRead(ns), nil
	}

	public.ID = uuid.New().String()
	public.CreatedAt = time.Now()

	manifest := CreateNamespaceManifest(public)
	res, err := client.K8sClient.CoreV1().Namespaces().Create(context.TODO(), manifest, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return models.CreateNamespacePublicFromRead(res), nil
}

func (client *Client) UpdateNamespace(public *models.NamespacePublic) (*models.NamespacePublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update namespace %s. details: %w", public.Name, err)
	}

	if public.ID == "" {
		log.Println("no id in namespace when updating. assuming it was deleted")
		return nil, nil
	}

	list, err := client.K8sClient.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", keys.ManifestLabelID, public.ID),
	})
	if err != nil {
		return nil, err
	}

	for _, item := range list.Items {
		if FindLabel(item.ObjectMeta.Labels, keys.ManifestLabelName, public.Name) {
			idLabel := GetLabel(item.ObjectMeta.Labels, keys.ManifestLabelID)
			if idLabel != "" {
				manifest := CreateNamespaceManifest(public)
				_, err = client.K8sClient.CoreV1().Namespaces().Update(context.TODO(), manifest, metav1.UpdateOptions{})
				if err != nil {
					return nil, makeError(err)
				}

				return models.CreateNamespacePublicFromRead(&item), nil
			}
		}
	}

	log.Println("namespace", public.Name, "not found when updating. assuming it was deleted")
	return nil, nil
}

func (client *Client) DeleteNamespace(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete namespace %s. details: %w", name, err)
	}

	list, err := client.K8sClient.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{
		TypeMeta:      metav1.TypeMeta{},
		LabelSelector: fmt.Sprintf("%s=%s", keys.ManifestLabelName, name),
		FieldSelector: "",
	})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		err = client.K8sClient.CoreV1().Namespaces().Delete(context.TODO(), item.Name, metav1.DeleteOptions{
			TypeMeta: metav1.TypeMeta{},
		})
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

package k8s

import (
	"context"
	"fmt"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"time"
)

// ReadAllNamespaces reads all namespaces from Kubernetes.
// If prefix is supplied, only namespaces with that prefix will be returned.
func (client *Client) ReadAllNamespaces(prefix *string) ([]models.NamespacePublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read namespaces. details: %w", err)
	}

	list, err := client.K8sClient.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, makeError(err)
	}

	var result []models.NamespacePublic
	for _, item := range list.Items {
		if prefix != nil && !strings.HasPrefix(item.ObjectMeta.Name, *prefix) {
			continue
		}

		result = append(result, *models.CreateNamespacePublicFromRead(&item))
	}

	return result, nil
}

// ReadNamespace reads a Namespace from Kubernetes.
func (client *Client) ReadNamespace(name string) (*models.NamespacePublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read namespace %s. details: %w", name, err)
	}

	if name == "" {
		return nil, nil
	}

	res, err := client.K8sClient.CoreV1().Namespaces().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	return models.CreateNamespacePublicFromRead(res), nil
}

// CreateNamespace creates a Namespace in Kubernetes.
func (client *Client) CreateNamespace(public *models.NamespacePublic) (*models.NamespacePublic, error) {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to create namespace %s. details: %w", public.Name, err)
	}

	ns, err := client.K8sClient.CoreV1().Namespaces().Get(context.TODO(), public.Name, metav1.GetOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return nil, makeErr(err)
	}

	if err == nil {
		return models.CreateNamespacePublicFromRead(ns), nil
	}

	public.CreatedAt = time.Now()

	manifest := CreateNamespaceManifest(public)
	res, err := client.K8sClient.CoreV1().Namespaces().Create(context.TODO(), manifest, metav1.CreateOptions{})
	if err != nil {
		return nil, makeErr(err)
	}

	return models.CreateNamespacePublicFromRead(res), nil
}

// UpdateNamespace updates a Namespace in Kubernetes.
func (client *Client) UpdateNamespace(public *models.NamespacePublic) (*models.NamespacePublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update namespace %s. details: %w", public.Name, err)
	}

	if public.Name == "" {
		log.Println("No id in namespace when updating. Assuming it was deleted")
		return nil, nil
	}

	ns, err := client.K8sClient.CoreV1().Namespaces().Update(context.TODO(), CreateNamespaceManifest(public), metav1.UpdateOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return nil, makeError(err)
	}

	if err == nil {
		return models.CreateNamespacePublicFromRead(ns), nil
	}

	log.Println("Namespace", public.Name, "not found when updating. Assuming it was deleted")
	return nil, nil
}

// DeleteNamespace deletes a Namespace in Kubernetes.
func (client *Client) DeleteNamespace(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete namespace %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("No name supplied when deleting namespace. Assuming it was deleted")
		return nil
	}

	err := client.K8sClient.CoreV1().Namespaces().Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return makeError(err)
	}

	err = client.waitNamespaceDeleted(name)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// DeleteNamespaceIfEmpty deletes a Namespace in Kubernetes if it is empty.
func (client *Client) DeleteNamespaceIfEmpty(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete namespace %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("No name supplied when deleting namespace. Assuming it was deleted")
		return nil
	}

	ns, err := client.K8sClient.CoreV1().Namespaces().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil
		}

		return makeError(err)
	}

	if len(ns.Status.Conditions) == 0 {
		return client.DeleteNamespace(name)
	}

	return nil
}

// waitNamespaceDeleted waits for a namespace to be deleted.
func (client *Client) waitNamespaceDeleted(name string) error {
	maxWait := 120
	for i := 0; i < maxWait; i++ {
		time.Sleep(1 * time.Second)
		_, err := client.K8sClient.CoreV1().Namespaces().Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil && IsNotFoundErr(err) {
			return nil
		}
	}

	return fmt.Errorf("timeout waiting for namespace %s to be deleted", name)
}

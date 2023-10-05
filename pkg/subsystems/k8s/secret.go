package k8s

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/pkg/subsystems/k8s/keys"
	"go-deploy/pkg/subsystems/k8s/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func (client *Client) ReadSecret(id string) (*models.SecretPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read deployment %s. details: %w", id, err)
	}

	if id == "" {
		return nil, nil
	}

	if client.Namespace == "" {
		return nil, nil
	}

	namespaceCreated, err := client.NamespaceCreated(client.Namespace)
	if err != nil {
		return nil, makeError(err)
	}

	if !namespaceCreated {
		return nil, makeError(fmt.Errorf("no such namespace %s", client.Namespace))
	}

	list, err := client.K8sClient.CoreV1().Secrets(client.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, item := range list.Items {
		idLabel := GetLabel(item.ObjectMeta.Labels, keys.ManifestLabelID)
		if idLabel == id {
			return models.CreateSecretPublicFromRead(&item), nil
		}
	}

	return nil, nil
}

func (client *Client) CreateSecret(public *models.SecretPublic) (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create deployment %s. details: %w", public.Name, err)
	}

	if public.Name == "" {
		return "", nil
	}

	if public.Namespace == "" {
		return "", nil
	}

	namespaceCreated, err := client.NamespaceCreated(public.Namespace)
	if err != nil {
		return "", makeError(err)
	}

	if !namespaceCreated {
		return "", makeError(fmt.Errorf("no such namespace %s", public.Namespace))
	}

	list, err := client.K8sClient.CoreV1().Secrets(public.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", makeError(err)
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
	public.CreatedAt = time.Now()

	manifest := CreateSecretManifest(public)
	result, err := client.K8sClient.CoreV1().Secrets(public.Namespace).Create(context.TODO(), manifest, metav1.CreateOptions{})
	if err != nil {
		return "", makeError(err)
	}

	return GetLabel(result.ObjectMeta.Labels, keys.ManifestLabelID), nil
}

func (client *Client) UpdateSecret(public *models.SecretPublic) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update deployment %s. details: %w", public.Name, err)
	}

	if public.Name == "" {
		return nil
	}

	if public.Namespace == "" {
		return nil
	}

	namespaceCreated, err := client.NamespaceCreated(public.Namespace)
	if err != nil {
		return makeError(err)
	}

	if !namespaceCreated {
		return makeError(fmt.Errorf("no such namespace %s", public.Namespace))
	}

	list, err := client.K8sClient.CoreV1().Secrets(public.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return makeError(err)
	}

	for _, item := range list.Items {
		if FindLabel(item.ObjectMeta.Labels, keys.ManifestLabelName, public.Name) {
			idLabel := GetLabel(item.ObjectMeta.Labels, keys.ManifestLabelID)
			if idLabel != "" {
				manifest := CreateSecretManifest(public)
				manifest.ObjectMeta.Labels[keys.ManifestLabelID] = idLabel
				_, err := client.K8sClient.CoreV1().Secrets(public.Namespace).Update(context.TODO(), manifest, metav1.UpdateOptions{})
				if err != nil {
					return makeError(err)
				}
				return nil
			}
		}
	}

	return nil
}

func (client *Client) DeleteSecret(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete deployment %s. details: %w", id, err)
	}

	if id == "" {
		return nil
	}

	if client.Namespace == "" {
		return nil
	}

	namespaceCreated, err := client.NamespaceCreated(client.Namespace)
	if err != nil {
		return makeError(err)
	}

	if !namespaceCreated {
		return makeError(fmt.Errorf("no such namespace %s", client.Namespace))
	}

	list, err := client.K8sClient.CoreV1().Secrets(client.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return makeError(err)
	}

	for _, item := range list.Items {
		idLabel := GetLabel(item.ObjectMeta.Labels, keys.ManifestLabelID)
		if idLabel == id {
			err := client.K8sClient.CoreV1().Secrets(client.Namespace).Delete(context.TODO(), item.Name, metav1.DeleteOptions{})
			if err != nil {
				return makeError(err)
			}
			return nil
		}
	}

	return nil
}

package k8s

import (
	"context"
	"fmt"
	"go-deploy/pkg/subsystems/k8s/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"time"
)

// ReadSecret reads a Secret from Kubernetes.
func (client *Client) ReadSecret(name string) (*models.SecretPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read k8s secret %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("no name supplied when reading k8s secret. assuming it was deleted")
		return nil, nil
	}

	res, err := client.K8sClient.CoreV1().Secrets(client.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	return models.CreateSecretPublicFromRead(res), nil
}

// CreateSecret creates a Secret in Kubernetes.
func (client *Client) CreateSecret(public *models.SecretPublic) (*models.SecretPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s secret %s. details: %w", public.Name, err)
	}

	secret, err := client.K8sClient.CoreV1().Secrets(public.Namespace).Get(context.TODO(), public.Name, metav1.GetOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return nil, makeError(err)
	}

	if err == nil {
		return models.CreateSecretPublicFromRead(secret), nil
	}

	public.CreatedAt = time.Now()

	manifest := CreateSecretManifest(public)
	res, err := client.K8sClient.CoreV1().Secrets(public.Namespace).Create(context.TODO(), manifest, metav1.CreateOptions{})
	if err != nil {
		return nil, makeError(err)
	}

	return models.CreateSecretPublicFromRead(res), nil
}

// UpdateSecret updates a Secret in Kubernetes.
func (client *Client) UpdateSecret(public *models.SecretPublic) (*models.SecretPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update k8s secret %s. details: %w", public.Name, err)
	}

	if public.Name == "" {
		log.Println("no name supplied when updating k8s secret. assuming it was deleted")
		return nil, nil
	}

	res, err := client.K8sClient.CoreV1().Secrets(public.Namespace).Update(context.TODO(), CreateSecretManifest(public), metav1.UpdateOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return nil, makeError(err)
	}

	if err == nil {
		return models.CreateSecretPublicFromRead(res), nil
	}

	log.Println("k8s secret", public.Name, "not found when updating. assuming it was deleted")
	return nil, nil
}

// DeleteSecret deletes a Secret in Kubernetes.
func (client *Client) DeleteSecret(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s secret %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("no name supplied when deleting k8s secret. assuming it was deleted")
		return nil
	}

	err := client.K8sClient.CoreV1().Secrets(client.Namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return makeError(err)
	}

	return nil
}

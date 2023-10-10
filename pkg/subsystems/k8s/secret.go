package k8s

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/pkg/subsystems/k8s/keys"
	"go-deploy/pkg/subsystems/k8s/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"time"
)

func (client *Client) ReadSecret(id string) (*models.SecretPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read k8s secret %s. details: %w", id, err)
	}

	if id == "" {
		log.Println("no id supplied when reading k8s secret. assuming it was deleted")
		return nil, nil
	}

	list, err := client.K8sClient.CoreV1().Secrets(client.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", keys.ManifestLabelID, id),
	})
	if err != nil {
		return nil, makeError(err)
	}

	if len(list.Items) > 0 {
		return models.CreateSecretPublicFromRead(&list.Items[0]), nil
	}

	return nil, nil
}

func (client *Client) CreateSecret(public *models.SecretPublic) (*models.SecretPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s secret %s. details: %w", public.Name, err)
	}

	secret, err := client.K8sClient.CoreV1().Secrets(public.Namespace).Get(context.TODO(), public.Name, metav1.GetOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return nil, makeError(err)
	}

	if secret != nil {
		return models.CreateSecretPublicFromRead(secret), nil
	}

	public.ID = uuid.New().String()
	public.CreatedAt = time.Now()

	manifest := CreateSecretManifest(public)
	res, err := client.K8sClient.CoreV1().Secrets(public.Namespace).Create(context.TODO(), manifest, metav1.CreateOptions{})
	if err != nil {
		return nil, makeError(err)
	}

	return models.CreateSecretPublicFromRead(res), nil
}

func (client *Client) UpdateSecret(public *models.SecretPublic) (*models.SecretPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update k8s secret %s. details: %w", public.Name, err)
	}

	if public.ID == "" {
		log.Println("no id supplied when updating k8s secret. assuming it was deleted")
		return nil, nil
	}

	res, err := client.K8sClient.CoreV1().Secrets(public.Namespace).Update(context.TODO(), CreateSecretManifest(public), metav1.UpdateOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return nil, makeError(err)
	}

	if res != nil {
		return models.CreateSecretPublicFromRead(res), nil
	}

	log.Println("k8s secret", public.Name, "not found when updating. assuming it was deleted")
	return nil, nil
}

func (client *Client) DeleteSecret(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s secret %s. details: %w", id, err)
	}

	if id == "" {
		log.Println("no id supplied when deleting k8s secret. assuming it was deleted")
		return nil
	}

	list, err := client.K8sClient.CoreV1().Secrets(client.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", keys.ManifestLabelID, id),
	})
	if err != nil {
		return makeError(err)
	}

	for _, secret := range list.Items {
		err = client.K8sClient.CoreV1().Secrets(client.Namespace).Delete(context.TODO(), secret.Name, metav1.DeleteOptions{})
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

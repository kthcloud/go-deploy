package k8s

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/pkg/subsystems/k8s/keys"
	"go-deploy/pkg/subsystems/k8s/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
)

func (client *Client) ReadPV(id string) (*models.PvPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read k8s persistent volume %s. details: %w", id, err)
	}

	if id == "" {
		return nil, nil
	}

	list, err := client.K8sClient.CoreV1().PersistentVolumes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, makeError(err)
	}

	for _, item := range list.Items {
		idLabel := GetLabel(item.ObjectMeta.Labels, keys.ManifestLabelID)
		if idLabel == id {
			return models.CreatePvPublicFromRead(&item), nil
		}
	}

	return nil, nil
}

func (client *Client) CreatePV(public *models.PvPublic) (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to k8s persistent volume %s. details: %w", public.Name, err)
	}

	if public.Name == "" {
		return "", nil
	}

	list, err := client.K8sClient.CoreV1().PersistentVolumes().List(context.TODO(), metav1.ListOptions{})
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
	manifest := CreatePvManifest(public)
	_, err = client.K8sClient.CoreV1().PersistentVolumes().Create(context.TODO(), manifest, metav1.CreateOptions{})
	if err != nil {
		return "", makeError(err)
	}

	return public.ID, nil
}

func (client *Client) DeletePV(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s persistent volume %s. details: %w", id, err)
	}

	if id == "" {
		return nil
	}

	list, err := client.K8sClient.CoreV1().PersistentVolumes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return makeError(err)
	}

	for _, item := range list.Items {
		idLabel := GetLabel(item.ObjectMeta.Labels, keys.ManifestLabelID)
		if idLabel == id {
			err = client.K8sClient.CoreV1().PersistentVolumes().Delete(context.TODO(), item.Name, metav1.DeleteOptions{})
			if err != nil {
				return makeError(err)
			}

			return nil
		}
	}

	log.Println("k8s persistent volume", id, "not found when deleting. assuming it was deleted")
	return nil
}

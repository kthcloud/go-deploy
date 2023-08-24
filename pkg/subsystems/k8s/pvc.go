package k8s

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/pkg/subsystems/k8s/keys"
	"go-deploy/pkg/subsystems/k8s/models"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
)

func (client *Client) ReadPVC(namespace string, id string) (*models.PvcPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read k8s persistent volume claim %s. details: %s", id, err)
	}

	if id == "" {
		return nil, nil
	}

	if namespace == "" {
		return nil, nil
	}

	list, err := client.K8sClient.CoreV1().PersistentVolumeClaims(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, makeError(err)
	}

	for _, item := range list.Items {
		idLabel := GetLabel(item.ObjectMeta.Labels, keys.ManifestLabelID)
		if idLabel == id {
			return models.CreatePvcPublicFromRead(&item), nil
		}
	}

	return nil, nil
}

func (client *Client) CreatePVC(public *models.PvcPublic) (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to k8s persistent volume claim %s. details: %s", public.Name, err)
	}

	if public.Name == "" {
		return "", nil
	}

	if public.Namespace == "" {
		return "", nil
	}

	list, err := client.K8sClient.CoreV1().PersistentVolumeClaims(public.Namespace).List(context.TODO(), v1.ListOptions{})
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
	manifest := CreatePvcManifest(public)
	_, err = client.K8sClient.CoreV1().PersistentVolumeClaims(public.Namespace).Create(context.TODO(), manifest, v1.CreateOptions{})
	if err != nil {
		return "", makeError(err)
	}

	return public.ID, nil
}

func (client *Client) DeletePVC(namespace, id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s persistent volume claim %s. details: %s", id, err)
	}

	if id == "" {
		return nil
	}

	if namespace == "" {
		return nil
	}

	list, err := client.K8sClient.CoreV1().PersistentVolumeClaims(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return makeError(err)
	}

	for _, item := range list.Items {
		idLabel := GetLabel(item.ObjectMeta.Labels, keys.ManifestLabelID)
		if idLabel == id {
			err = client.K8sClient.CoreV1().PersistentVolumeClaims(item.Namespace).Delete(context.TODO(), item.Name, v1.DeleteOptions{})
			if err != nil {
				return makeError(err)
			}

			return nil
		}
	}

	log.Println("k8s persistent volume claim", id, "not found when deleting. assuming it was deleted")
	return nil
}

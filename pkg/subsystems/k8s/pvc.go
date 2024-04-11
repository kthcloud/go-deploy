package k8s

import (
	"context"
	"fmt"
	"go-deploy/pkg/log"
	"go-deploy/pkg/subsystems/k8s/models"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// ReadPVC reads a PersistentVolumeClaim from Kubernetes.
func (client *Client) ReadPVC(name string) (*models.PvcPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read k8s pvc %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("No name supplied when reading k8s pvc. Assuming it was deleted")
		return nil, nil
	}

	res, err := client.K8sClient.CoreV1().PersistentVolumeClaims(client.Namespace).Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	return models.CreatePvcPublicFromRead(res), nil
}

// CreatePVC creates a PersistentVolumeClaim in Kubernetes.
func (client *Client) CreatePVC(public *models.PvcPublic) (*models.PvcPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s pvc %s. details: %w", public.Name, err)
	}

	if public.Name == "" {
		log.Println("No name supplied when creating k8s pvc. Assuming it was deleted")
		return nil, nil
	}

	pvc, err := client.K8sClient.CoreV1().PersistentVolumeClaims(public.Namespace).Get(context.TODO(), public.Name, v1.GetOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return nil, makeError(err)
	}

	if err == nil {
		return models.CreatePvcPublicFromRead(pvc), nil
	}

	public.CreatedAt = time.Now()

	manifest := CreatePvcManifest(public)
	res, err := client.K8sClient.CoreV1().PersistentVolumeClaims(public.Namespace).Create(context.TODO(), manifest, v1.CreateOptions{})
	if err != nil {
		return nil, makeError(err)
	}

	return models.CreatePvcPublicFromRead(res), nil
}

// DeletePVC deletes a PersistentVolumeClaim in Kubernetes.
func (client *Client) DeletePVC(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s pvc %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("No name supplied when deleting k8s pvc. Assuming it was deleted")
		return nil
	}

	err := client.K8sClient.CoreV1().PersistentVolumeClaims(client.Namespace).Delete(context.TODO(), name, v1.DeleteOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return makeError(err)
	}

	err = client.waitPvcDeleted(name)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// waitPvcDeleted waits for a PersistentVolumeClaim to be deleted.
func (client *Client) waitPvcDeleted(name string) error {
	maxWait := 120
	for i := 0; i < maxWait; i++ {
		time.Sleep(1 * time.Second)
		_, err := client.K8sClient.CoreV1().PersistentVolumeClaims(client.Namespace).Get(context.TODO(), name, v1.GetOptions{})
		if err != nil && IsNotFoundErr(err) {
			return nil
		}
	}

	return fmt.Errorf("timeout waiting for pvc %s to be deleted", name)
}

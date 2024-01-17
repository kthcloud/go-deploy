package k8s

import (
	"context"
	"fmt"
	"go-deploy/pkg/subsystems/k8s/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"time"
)

// ReadPV reads a PersistentVolume from Kubernetes.
func (client *Client) ReadPV(name string) (*models.PvPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read k8s pv %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("no name supplied when reading k8s pv. assuming it was deleted")
		return nil, nil
	}

	res, err := client.K8sClient.CoreV1().PersistentVolumes().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	return models.CreatePvPublicFromRead(res), nil
}

// CreatePV creates a PersistentVolume in Kubernetes.
func (client *Client) CreatePV(public *models.PvPublic) (*models.PvPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s pv %s. details: %w", public.Name, err)
	}

	pv, err := client.K8sClient.CoreV1().PersistentVolumes().Get(context.TODO(), public.Name, metav1.GetOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return nil, makeError(err)
	}

	if err == nil {
		return models.CreatePvPublicFromRead(pv), nil
	}

	public.CreatedAt = time.Now()

	manifest := CreatePvManifest(public)
	res, err := client.K8sClient.CoreV1().PersistentVolumes().Create(context.TODO(), manifest, metav1.CreateOptions{})
	if err != nil {
		return nil, makeError(err)
	}

	return models.CreatePvPublicFromRead(res), nil
}

// DeletePV deletes a PersistentVolume in Kubernetes.
func (client *Client) DeletePV(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s pv %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("no name supplied when deleting k8s pv. assuming it was deleted")
		return nil
	}

	err := client.K8sClient.CoreV1().PersistentVolumes().Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return makeError(err)
	}
	
	err = client.waitPvDeleted(name)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// waitPvDeleted waits for a PersistentVolume to be deleted.
func (client *Client) waitPvDeleted(name string) error {
	maxWait := 120
	for i := 0; i < maxWait; i++ {
		time.Sleep(1 * time.Second)
		_, err := client.K8sClient.CoreV1().PersistentVolumes().Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil && IsNotFoundErr(err) {
			return nil
		}
	}

	return fmt.Errorf("timeout waiting for pv %s to be deleted", name)
}

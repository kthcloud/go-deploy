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

// ReadService reads a Service from Kubernetes.
func (client *Client) ReadService(id string) (*models.ServicePublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read k8s service %s. details: %w", id, err)
	}

	if id == "" {
		log.Println("no id supplied when reading k8s service. assuming it was deleted")
		return nil, nil
	}

	list, err := client.K8sClient.CoreV1().Services(client.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", keys.ManifestLabelID, id),
	})
	if err != nil {
		return nil, makeError(err)
	}

	if len(list.Items) > 0 {
		return models.CreateServicePublicFromRead(&list.Items[0]), nil
	}

	return nil, nil
}

// CreateService creates a Service in Kubernetes.
func (client *Client) CreateService(public *models.ServicePublic) (*models.ServicePublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s service %s. details: %w", public.Name, err)
	}

	service, err := client.K8sClient.CoreV1().Services(public.Namespace).Get(context.TODO(), public.Name, metav1.GetOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return nil, makeError(err)
	}

	if err == nil {
		return models.CreateServicePublicFromRead(service), nil
	}

	public.ID = uuid.New().String()
	public.CreatedAt = time.Now()

	manifest := CreateServiceManifest(public)
	res, err := client.K8sClient.CoreV1().Services(public.Namespace).Create(context.TODO(), manifest, metav1.CreateOptions{})
	if err != nil {
		return nil, makeError(err)
	}

	return models.CreateServicePublicFromRead(res), nil
}

// UpdateService updates a Service in Kubernetes.
func (client *Client) UpdateService(public *models.ServicePublic) (*models.ServicePublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update k8s service %s. details: %w", public.ID, err)
	}

	if public.ID == "" {
		log.Println("no id supplied when updating k8s service. assuming it was deleted")
		return nil, nil
	}

	res, err := client.K8sClient.CoreV1().Services(public.Namespace).Update(context.TODO(), CreateServiceManifest(public), metav1.UpdateOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return nil, makeError(err)
	}

	if err == nil {
		return models.CreateServicePublicFromRead(res), nil
	}

	log.Println("k8s service", public.Name, "not found when updating. assuming it was deleted")
	return nil, nil
}

// DeleteService deletes a Service in Kubernetes.
func (client *Client) DeleteService(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s service %s. details: %w", id, err)
	}

	if id == "" {
		log.Println("no id supplied when deleting k8s service. assuming it was deleted")
		return nil
	}

	list, err := client.K8sClient.CoreV1().Services(client.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", keys.ManifestLabelID, id),
	})
	if err != nil {
		return makeError(err)
	}

	for _, item := range list.Items {
		err = client.K8sClient.CoreV1().Services(client.Namespace).Delete(context.TODO(), item.Name, metav1.DeleteOptions{})
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

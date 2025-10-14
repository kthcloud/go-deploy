package k8s

import (
	"context"
	"fmt"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// ReadService reads a Service from Kubernetes.
func (client *Client) ReadService(name string) (*models.ServicePublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read k8s service %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("No name supplied when reading k8s service. Assuming it was deleted")
		return nil, nil
	}

	res, err := client.K8sClient.CoreV1().Services(client.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	return models.CreateServicePublicFromRead(res), nil
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
		return fmt.Errorf("failed to update k8s service %s. details: %w", public.Name, err)
	}

	if public.Name == "" {
		log.Println("No name supplied when updating k8s service. Assuming it was deleted")
		return nil, nil
	}

	res, err := client.K8sClient.CoreV1().Services(public.Namespace).Update(context.TODO(), CreateServiceManifest(public), metav1.UpdateOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return nil, makeError(err)
	}

	if err == nil {
		return models.CreateServicePublicFromRead(res), nil
	}

	log.Println("K8s service", public.Name, "not found when updating. Assuming it was deleted")
	return nil, nil
}

// DeleteService deletes a Service in Kubernetes.
func (client *Client) DeleteService(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s service %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("No name supplied when deleting k8s service. Assuming it was deleted")
		return nil
	}

	err := client.K8sClient.CoreV1().Services(client.Namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return makeError(err)
	}

	return nil
}

package k8s

import (
	"context"
	"fmt"
	"go-deploy/pkg/log"
	"go-deploy/pkg/subsystems/k8s/errors"
	"go-deploy/pkg/subsystems/k8s/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"time"
)

// ReadIngress reads a Ingress from Kubernetes.
func (client *Client) ReadIngress(name string) (*models.IngressPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read k8s ingress %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("No name supplied when reading k8s ingress. Assuming it was deleted")
		return nil, nil
	}

	res, err := client.K8sClient.NetworkingV1().Ingresses(client.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	return models.CreateIngressPublicFromRead(res), nil
}

// CreateIngress creates a Ingress in Kubernetes.
func (client *Client) CreateIngress(public *models.IngressPublic) (*models.IngressPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s ingress %s. details: %w", public.Name, err)
	}

	ingress, err := client.K8sClient.NetworkingV1().Ingresses(public.Namespace).Get(context.TODO(), public.Name, metav1.GetOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return nil, makeError(err)
	}

	if err == nil {
		return models.CreateIngressPublicFromRead(ingress), nil
	}

	public.CreatedAt = time.Now()

	manifest := CreateIngressManifest(public)
	res, err := client.K8sClient.NetworkingV1().Ingresses(public.Namespace).Create(context.TODO(), manifest, metav1.CreateOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "is already defined in ingress") {
			return nil, makeError(errors.IngressHostInUseErr)
		}

		return nil, makeError(err)
	}

	return models.CreateIngressPublicFromRead(res), nil
}

// UpdateIngress updates a Ingress in Kubernetes.
func (client *Client) UpdateIngress(public *models.IngressPublic) (*models.IngressPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update k8s ingress %s. details: %w", public.Name, err)
	}

	if public.Name == "" {
		log.Println("No name supplied when updating k8s ingress. Assuming it was deleted")
		return nil, nil
	}

	ingress, err := client.K8sClient.NetworkingV1().Ingresses(public.Namespace).Update(context.TODO(), CreateIngressManifest(public), metav1.UpdateOptions{})
	if err != nil && !IsNotFoundErr(err) {
		if strings.Contains(err.Error(), "is already defined in ingress") {
			return nil, makeError(errors.IngressHostInUseErr)
		}

		return nil, makeError(err)
	}

	if err == nil {
		return models.CreateIngressPublicFromRead(ingress), nil
	}

	log.Println("K8s ingress", public.Name, "not found when updating. Assuming it was deleted")
	return nil, nil
}

// DeleteIngress deletes a Ingress in Kubernetes.
func (client *Client) DeleteIngress(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s ingress %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("No name supplied when deleting k8s ingress. Assuming it was deleted")
		return nil
	}

	err := client.K8sClient.NetworkingV1().Ingresses(client.Namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return makeError(err)
	}

	return nil
}

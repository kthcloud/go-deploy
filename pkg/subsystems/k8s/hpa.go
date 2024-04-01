package k8s

import (
	"context"
	"fmt"
	"go-deploy/pkg/log"
	"go-deploy/pkg/subsystems/k8s/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// ReadHPA reads a HorizontalPodAutoscaler from Kubernetes.
func (client *Client) ReadHPA(name string) (*models.HpaPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read k8s hpa %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("no name supplied when reading k8s hpa. assuming it was deleted")
		return nil, nil
	}

	res, err := client.K8sClient.AutoscalingV2().HorizontalPodAutoscalers(client.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	return models.CreateHpaPublicFromRead(res), nil
}

// CreateHPA creates a HorizontalPodAutoscaler in Kubernetes.
func (client *Client) CreateHPA(public *models.HpaPublic) (*models.HpaPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s hpa %s. details: %w", public.Name, err)
	}

	hpa, err := client.K8sClient.AutoscalingV2().HorizontalPodAutoscalers(public.Namespace).Get(context.TODO(), public.Name, metav1.GetOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return nil, makeError(err)
	}

	if err == nil {
		return models.CreateHpaPublicFromRead(hpa), nil
	}

	public.CreatedAt = time.Now()

	manifest := CreateHpaManifest(public)
	res, err := client.K8sClient.AutoscalingV2().HorizontalPodAutoscalers(public.Namespace).Create(context.TODO(), manifest, metav1.CreateOptions{})
	if err != nil {
		return nil, makeError(err)
	}

	return models.CreateHpaPublicFromRead(res), nil
}

// UpdateHPA updates a HorizontalPodAutoscaler in Kubernetes.
func (client *Client) UpdateHPA(public *models.HpaPublic) (*models.HpaPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update k8s hpa %s. details: %w", public.Name, err)
	}

	if public.Name == "" {
		log.Println("no name supplied when updating k8s hpa. assuming it was deleted")
		return nil, nil
	}

	hpa, err := client.K8sClient.AutoscalingV2().HorizontalPodAutoscalers(public.Namespace).Update(context.TODO(), CreateHpaManifest(public), metav1.UpdateOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return nil, makeError(err)
	}

	if err == nil {
		return models.CreateHpaPublicFromRead(hpa), nil
	}

	log.Println("k8s hpa", public.Name, "not found when updating. assuming it was deleted")
	return nil, nil
}

// DeleteHPA deletes a HorizontalPodAutoscaler in Kubernetes.
func (client *Client) DeleteHPA(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s hpa %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("no name supplied when deleting k8s hpa. assuming it was deleted")
		return nil
	}

	err := client.K8sClient.AutoscalingV2().HorizontalPodAutoscalers(client.Namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return makeError(err)
	}

	return nil
}

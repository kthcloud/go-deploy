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

// ReadHPA reads a HorizontalPodAutoscaler from Kubernetes.
func (client *Client) ReadHPA(id string) (*models.HpaPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read k8s hpa %s. details: %w", id, err)
	}

	if id == "" {
		log.Println("no id supplied when reading k8s hpa. assuming it was deleted")
		return nil, nil
	}

	list, err := client.K8sClient.AutoscalingV2().HorizontalPodAutoscalers(client.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", keys.ManifestLabelID, id),
	})
	if err != nil {
		return nil, makeError(err)
	}

	if len(list.Items) > 0 {
		return models.CreateHpaPublicFromRead(&list.Items[0]), nil
	}

	return nil, nil
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

	public.ID = uuid.New().String()
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
		return fmt.Errorf("failed to update k8s hpa %s. details: %w", public.ID, err)
	}

	if public.ID == "" {
		log.Println("no id supplied when updating k8s hpa. assuming it was deleted")
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
func (client *Client) DeleteHPA(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s hpa %s. details: %w", id, err)
	}

	if id == "" {
		log.Println("no id supplied when deleting k8s hpa. assuming it was deleted")
		return nil
	}

	list, err := client.K8sClient.AutoscalingV2().HorizontalPodAutoscalers(client.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", keys.ManifestLabelID, id),
	})
	if err != nil {
		return makeError(err)
	}

	for _, hpa := range list.Items {
		err = client.K8sClient.AutoscalingV2().HorizontalPodAutoscalers(client.Namespace).Delete(context.TODO(), hpa.Name, metav1.DeleteOptions{})
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

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

func (client *Client) ReadIngress(namespace, id string) (*models.IngressPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read deployment %s. details: %w", id, err)
	}

	if id == "" {
		return nil, nil
	}

	if namespace == "" {
		return nil, nil
	}

	namespaceCreated, err := client.NamespaceCreated(namespace)
	if err != nil {
		return nil, makeError(err)
	}

	if !namespaceCreated {
		return nil, makeError(fmt.Errorf("no such namespace %s", namespace))
	}

	list, err := client.K8sClient.NetworkingV1().Ingresses(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, item := range list.Items {
		idLabel := GetLabel(item.ObjectMeta.Labels, keys.ManifestLabelID)
		if idLabel == id {
			return models.CreateIngressPublicFromRead(&item), nil
		}
	}

	return nil, nil
}

func (client *Client) CreateIngress(public *models.IngressPublic) (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to k8s ingress %s. details: %w", public.Name, err)
	}

	if public.Name == "" {
		return "", nil
	}

	if public.Namespace == "" {
		return "", nil
	}

	namespaceCreated, err := client.NamespaceCreated(public.Namespace)
	if err != nil {
		return "", makeError(err)
	}

	if !namespaceCreated {
		return "", makeError(fmt.Errorf("no such namespace %s", public.Namespace))
	}

	list, err := client.K8sClient.NetworkingV1().Ingresses(public.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", err
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
	manifest := CreateIngressManifest(public)
	_, err = client.K8sClient.NetworkingV1().Ingresses(public.Namespace).Create(context.TODO(), manifest, metav1.CreateOptions{})
	if err != nil {
		return "", makeError(err)
	}

	return public.ID, nil
}

func (client *Client) UpdateIngress(public *models.IngressPublic) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update k8s ingress %s. details: %w", public.Name, err)
	}

	if public.ID == "" {
		return nil
	}

	if public.Namespace == "" {
		return nil
	}

	namespaceCreated, err := client.NamespaceCreated(public.Namespace)
	if err != nil {
		return makeError(err)
	}

	if !namespaceCreated {
		return makeError(fmt.Errorf("no such namespace %s", public.Namespace))
	}

	list, err := client.K8sClient.NetworkingV1().Ingresses(public.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		idLabel := GetLabel(item.ObjectMeta.Labels, keys.ManifestLabelID)
		if idLabel == public.ID {
			manifest := CreateIngressManifest(public)
			_, err = client.K8sClient.NetworkingV1().Ingresses(public.Namespace).Update(context.TODO(), manifest, metav1.UpdateOptions{})
			if err != nil {
				return makeError(err)
			}

			return nil
		}
	}

	log.Println("k8s ingress", public.Name, "not found when updating. assuming it was deleted")
	return nil
}

func (client *Client) DeleteIngress(namespace, id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s ingress %s. details: %w", id, err)
	}

	if id == "" {
		return nil
	}

	if namespace == "" {
		return nil
	}

	namespaceCreated, err := client.NamespaceCreated(namespace)
	if err != nil {
		return makeError(err)
	}

	if !namespaceCreated {
		return nil
	}

	list, err := client.K8sClient.NetworkingV1().Ingresses(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		idLabel := GetLabel(item.ObjectMeta.Labels, keys.ManifestLabelID)
		if idLabel == id {
			err = client.K8sClient.NetworkingV1().Ingresses(namespace).Delete(context.TODO(), item.Name, metav1.DeleteOptions{})
			if err != nil {
				return makeError(err)
			}

			return nil
		}
	}

	log.Println("k8s ingress", id, "not found when deleting. assuming it was deleted")
	return nil
}

package k8s

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/pkg/subsystems/k8s/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func (client *Client) ReadService(namespace, id string) (*models.ServicePublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read deployment %s. details: %s", id, err)
	}

	if id == "" {
		return nil, makeError(errors.New("id required"))
	}

	if namespace == "" {
		return nil, makeError(errors.New("namespace required"))
	}

	namespaceCreated, err := client.NamespaceCreated(namespace)
	if err != nil {
		return nil, makeError(err)
	}

	if !namespaceCreated {
		return nil, makeError(fmt.Errorf("no such namespace %s", namespace))
	}

	list, err := client.K8sClient.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, item := range list.Items {
		if strings.HasSuffix(item.Name, id) && len(item.Name) > len(uuid.New().String())+1 {
			return models.CreateServicePublicFromRead(&item), nil
		}
	}

	return nil, nil
}

func (client *Client) CreateService(public *models.ServicePublic) (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to k8s service %s. details: %s", public.Name, err)
	}

	if public.Name == "" {
		return "", makeError(errors.New("name required"))
	}

	if public.Namespace == "" {
		return "", makeError(errors.New("namespace required"))
	}

	namespaceCreated, err := client.NamespaceCreated(public.Namespace)
	if err != nil {
		return "", makeError(err)
	}

	if !namespaceCreated {
		return "", makeError(fmt.Errorf("no such namespace %s", public.Namespace))
	}

	list, err := client.K8sClient.CoreV1().Services(public.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	for _, item := range list.Items {
		if strings.HasPrefix(item.Name, public.Name) {
			idAndName, err := models.GetIdAndName(item.Name)
			if err == nil {
				return idAndName[0], nil
			}
		}
	}

	public.ID = uuid.New().String()
	manifest := CreateServiceManifest(public)
	_, err = client.K8sClient.CoreV1().Services(public.Namespace).Create(context.TODO(), manifest, metav1.CreateOptions{})
	if err != nil {
		return "", makeError(err)
	}

	return public.ID, nil
}

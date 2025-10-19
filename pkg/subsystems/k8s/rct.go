package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/models"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReadRCT reads a ResourceClaimTemplate from Kubernetes.
func (client *Client) ReadRCT(name string) (*models.ResourceClaimTemplatePublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read k8s RCT %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("No name supplied when reading k8s RCT. Assuming it was deleted")
		return nil, nil
	}

	res, err := client.K8sClient.ResourceV1().ResourceClaimTemplates(client.Namespace).Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	return models.CreateResourceClaimTemplatePublicFromRead(res), nil
}

// CreateRCT creates a ResourceClaimTemplate in Kubernetes.
func (client *Client) CreateRCT(public *models.ResourceClaimTemplatePublic) (*models.ResourceClaimTemplatePublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s RCT %s. details: %w", public.Name, err)
	}

	if public.Name == "" {
		log.Println("No name supplied when creating k8s RCT. Assuming it was deleted")
		return nil, nil
	}

	RCT, err := client.K8sClient.ResourceV1().ResourceClaimTemplates(public.Namespace).Get(context.TODO(), public.Name, v1.GetOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return nil, makeError(err)
	}

	if err == nil {
		return models.CreateResourceClaimTemplatePublicFromRead(RCT), nil
	}

	public.CreatedAt = time.Now()

	manifest := CreateResourceClaimTemplateManifest(public)
	res, err := client.K8sClient.ResourceV1().ResourceClaimTemplates(public.Namespace).Create(context.TODO(), manifest, v1.CreateOptions{})
	if err != nil {
		return nil, makeError(err)
	}

	return models.CreateResourceClaimTemplatePublicFromRead(res), nil
}

// DeleteRCT deletes a ResourceClaimTemplate in Kubernetes.
func (client *Client) DeleteRCT(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s RCT %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("No name supplied when deleting k8s RCT. Assuming it was deleted")
		return nil
	}

	err := client.K8sClient.ResourceV1().ResourceClaimTemplates(client.Namespace).Delete(context.TODO(), name, v1.DeleteOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return makeError(err)
	}

	err = client.waitRCTDeleted(name)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// waitRCTDeleted waits for a ResourceClaimTemplate to be deleted.
func (client *Client) waitRCTDeleted(name string) error {
	maxWait := 120
	for range maxWait {
		time.Sleep(1 * time.Second)
		_, err := client.K8sClient.ResourceV1().ResourceClaimTemplates(client.Namespace).Get(context.TODO(), name, v1.GetOptions{})
		if err != nil && IsNotFoundErr(err) {
			return nil
		}
	}

	return fmt.Errorf("timeout waiting for RCT %s to be deleted", name)
}

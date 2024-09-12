package k8s

import (
	"context"
	"fmt"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func (client *Client) ReadNetworkPolicy(name string) (*models.NetworkPolicyPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read k8s network policy %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("No name supplied when reading k8s network policy. Assuming it was deleted")
		return nil, nil
	}

	res, err := client.K8sClient.NetworkingV1().NetworkPolicies(client.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	return models.CreateNetworkPolicyPublicFromRead(res), nil
}

func (client *Client) CreateNetworkPolicy(public *models.NetworkPolicyPublic) (*models.NetworkPolicyPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s network policy %s. details: %w", public.Name, err)
	}

	policy, err := client.K8sClient.NetworkingV1().NetworkPolicies(public.Namespace).Get(context.TODO(), public.Name, metav1.GetOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return nil, makeError(err)
	}

	if err == nil {
		return models.CreateNetworkPolicyPublicFromRead(policy), nil
	}

	public.CreatedAt = time.Now()

	manifest := CreateNetworkPolicyManifest(public)
	res, err := client.K8sClient.NetworkingV1().NetworkPolicies(public.Namespace).Create(context.TODO(), manifest, metav1.CreateOptions{})
	if err != nil {
		return nil, makeError(err)
	}

	return models.CreateNetworkPolicyPublicFromRead(res), nil
}

func (client *Client) UpdateNetworkPolicy(public *models.NetworkPolicyPublic) (*models.NetworkPolicyPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update k8s network policy %s. details: %w", public.Name, err)
	}

	if public.Name == "" {
		log.Println("No name supplied when updating k8s network policy. Assuming it was deleted")
		return nil, nil
	}

	res, err := client.K8sClient.NetworkingV1().NetworkPolicies(public.Namespace).Update(context.TODO(), CreateNetworkPolicyManifest(public), metav1.UpdateOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return nil, makeError(err)
	}

	if err == nil {
		return models.CreateNetworkPolicyPublicFromRead(res), nil
	}

	log.Println("K8s network policy", public.Name, "not found when updating. Assuming it was deleted")
	return nil, nil
}

func (client *Client) DeleteNetworkPolicy(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s network policy %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("No name supplied when deleting k8s network policy. Assuming it was deleted")
		return nil
	}

	err := client.K8sClient.NetworkingV1().NetworkPolicies(client.Namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return makeError(err)
	}

	return nil
}

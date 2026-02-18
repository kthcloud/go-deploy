package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/models"
	k8sModels "github.com/kthcloud/go-deploy/pkg/subsystems/k8s/models"
	resourcev1 "k8s.io/api/resource/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

// ReadResourceClaim reads a ResourceClaim from Kubernetes.
func (client *Client) ReadResourceClaim(name string) (*models.ResourceClaimPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read k8s ResourceClaim %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("No name supplied when reading k8s ResourceClaim. Assuming it was deleted")
		return nil, nil
	}

	res, err := client.K8sClient.ResourceV1().ResourceClaims(client.Namespace).Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	return models.CreateResourceClaimPublicFromRead(res), nil
}

// CreateResourceClaim creates a ResourceClaim in Kubernetes.
func (client *Client) CreateResourceClaim(public *models.ResourceClaimPublic) (*models.ResourceClaimPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s ResourceClaim %s. details: %w", public.Name, err)
	}

	if public.Name == "" {
		log.Println("No name supplied when creating k8s ResourceClaim. Assuming it was deleted")
		return nil, nil
	}

	rc, err := client.K8sClient.ResourceV1().ResourceClaims(public.Namespace).Get(context.TODO(), public.Name, v1.GetOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return nil, makeError(err)
	}

	if err == nil {
		return models.CreateResourceClaimPublicFromRead(rc), nil
	}

	public.CreatedAt = time.Now()

	manifest := CreateResourceClaimManifest(public)
	res, err := client.K8sClient.ResourceV1().ResourceClaims(public.Namespace).Create(context.TODO(), manifest, v1.CreateOptions{})
	if err != nil {
		return nil, makeError(err)
	}

	return models.CreateResourceClaimPublicFromRead(res), nil
}

// DeleteResourceClaim deletes a ResourceClaim in Kubernetes.
func (client *Client) DeleteResourceClaim(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s ResourceClaim %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("No name supplied when deleting k8s ResourceClaim. Assuming it was deleted")
		return nil
	}

	err := client.K8sClient.ResourceV1().ResourceClaims(client.Namespace).Delete(context.TODO(), name, v1.DeleteOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return makeError(err)
	}

	err = client.waitResourceClaimDeleted(name)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// waitResourceClaimDeleted waits for a ResourceClaim to be deleted.
func (client *Client) waitResourceClaimDeleted(name string) error {
	maxWait := 120
	for range maxWait {
		time.Sleep(1 * time.Second)
		_, err := client.K8sClient.ResourceV1().ResourceClaims(client.Namespace).Get(context.TODO(), name, v1.GetOptions{})
		if err != nil && IsNotFoundErr(err) {
			return nil
		}
	}

	return fmt.Errorf("timeout waiting for ResourceClaim %s to be deleted", name)
}

// SetupResourceClaimWatcher is a function that sets up a resourceClaim watcher with a callback.
// It triggers the callback when a resourceClaim event occurs.
func (client *Client) SetupResourceClaimWatcher(ctx context.Context, callback func(resourceClaimName string, status k8sModels.ResourceClaimStatus, event string)) error {
	factory := informers.NewSharedInformerFactoryWithOptions(client.K8sClient, 0, informers.WithNamespace(client.Namespace))
	resoureClaimInformer := factory.Resource().V1().ResourceClaims().Informer()

	allowedResourceClaim := func(_ *resourcev1.ResourceClaim) bool {
		return true
	}

	_, err := resoureClaimInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			claim, ok := obj.(*resourcev1.ResourceClaim)
			if !ok {
				return
			}

			if !allowedResourceClaim(claim) {
				return
			}

			callback(claim.Name, rcToK8sModel(claim), PodEventAdded)
		},
		UpdateFunc: func(oldObj, newObj any) {
			claim, ok := newObj.(*resourcev1.ResourceClaim)
			if !ok {
				return
			}

			if !allowedResourceClaim(claim) {
				return
			}

			callback(claim.Name, rcToK8sModel(claim), PodEventUpdated)
		},
		DeleteFunc: func(obj any) {
			claim, ok := obj.(*resourcev1.ResourceClaim)
			if !ok {
				return
			}

			if !allowedResourceClaim(claim) {
				return
			}

			callback(claim.Name, rcToK8sModel(claim), PodEventDeleted)
		},
	})
	if err != nil {
		return err
	}

	factory.Start(ctx.Done())
	factory.WaitForCacheSync(ctx.Done())

	return nil
}

func rcToK8sModel(rc *resourcev1.ResourceClaim) k8sModels.ResourceClaimStatus {
	rcp := k8sModels.CreateResourceClaimPublicFromRead(rc)

	return k8sModels.ResourceClaimStatus{
		Allocated:         rcp.Allocated,
		AllocationResults: rcp.AllocationResults,
		Consumers:         rcp.Consumers,
	}
}

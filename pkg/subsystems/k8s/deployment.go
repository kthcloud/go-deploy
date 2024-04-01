package k8s

import (
	"context"
	"fmt"
	"go-deploy/pkg/log"
	"go-deploy/pkg/subsystems/k8s/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

// ReadDeployment reads a deployment from Kubernetes.
func (client *Client) ReadDeployment(name string) (*models.DeploymentPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read k8s deployment %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("no name supplied when reading k8s deployment. assuming it was deleted")
		return nil, nil
	}

	res, err := client.K8sClient.AppsV1().Deployments(client.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	return models.CreateDeploymentPublicFromRead(res), nil
}

// CreateDeployment creates a deployment in Kubernetes.
func (client *Client) CreateDeployment(public *models.DeploymentPublic) (*models.DeploymentPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create deployment %s. details: %w", public.Name, err)
	}

	deployment, err := client.K8sClient.AppsV1().Deployments(public.Namespace).Get(context.TODO(), public.Name, metav1.GetOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return nil, makeError(err)
	}

	if err == nil {
		return models.CreateDeploymentPublicFromRead(deployment), nil
	}

	public.CreatedAt = time.Now()

	manifest := CreateDeploymentManifest(public)
	res, err := client.K8sClient.AppsV1().Deployments(public.Namespace).Create(context.TODO(), manifest, metav1.CreateOptions{})
	if err != nil {
		return nil, makeError(err)
	}

	return models.CreateDeploymentPublicFromRead(res), nil
}

// UpdateDeployment updates a deployment in Kubernetes.
func (client *Client) UpdateDeployment(public *models.DeploymentPublic) (*models.DeploymentPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update k8s deployment %s. details: %w", public.Name, err)
	}

	if public.Name == "" {
		log.Println("no name supplied when updating k8s deployment. assuming it was deleted")
		return nil, nil
	}

	res, err := client.K8sClient.AppsV1().Deployments(public.Namespace).Update(context.TODO(), CreateDeploymentManifest(public), metav1.UpdateOptions{})
	if err != nil && !IsNotFoundErr(err) {

		// We treat immutability errors as if no update was done.
		// This is done because it is compatible with the repair pipeline,
		// which will try to update the deployment if it is not in a good state.
		// And if the update did not work, it will try to recreate -> which
		// will most likely solve the immutability issue.
		//
		// It is currently being tested, whether it's a good approach.
		// If so, it will be added to the other Kubernetes resources.
		if IsImmutabilityErr(err) {
			log.Println("k8s deployment", public.Name, "could not be updated due to immutability error. assuming bad state")
			return nil, nil
		}

		return nil, makeError(err)
	}

	if err == nil {
		return models.CreateDeploymentPublicFromRead(res), nil
	}

	log.Println("k8s deployment", public.Name, "not found when updating. assuming it was deleted")
	return nil, nil
}

// DeleteDeployment deletes a deployment in Kubernetes.
func (client *Client) DeleteDeployment(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete deployment %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("no name supplied when deleting k8s deployment. assuming it was deleted")
		return nil
	}

	err := client.K8sClient.AppsV1().Deployments(client.Namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return makeError(err)
	}

	return nil
}

// RestartDeployment restarts a deployment in Kubernetes.
// This is done by updating the deployment's annotation `kubectl.kubernetes.io/restartedAt` to the current time.
func (client *Client) RestartDeployment(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to restart k8s deployment %s. details: %w", name, err)
	}

	deployment, err := client.ReadDeployment(name)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		log.Println("no deployment found when restarting. assuming it was deleted")
		return nil
	}

	req := client.K8sClient.AppsV1().Deployments(client.Namespace)
	data := fmt.Sprintf(`{"spec": {"template": {"metadata": {"annotations": {"kubectl.kubernetes.io/restartedAt": "%s"}}}}}`, time.Now().Format("20060102150405"))

	_, err = req.Patch(context.TODO(), deployment.Name, types.StrategicMergePatchType, []byte(data), metav1.PatchOptions{})
	if err != nil {
		return makeError(err)
	}

	return nil
}

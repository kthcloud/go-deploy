package k8s

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/pkg/subsystems/k8s/keys"
	"go-deploy/pkg/subsystems/k8s/models"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"log"
	"time"
)

func (client *Client) createDeploymentWatcher(ctx context.Context, namespace, deployment string) (watch.Interface, error) {
	labelSelector := fmt.Sprintf("%s=%s", "app.kubernetes.io/name", deployment)

	opts := metav1.ListOptions{
		TypeMeta:      metav1.TypeMeta{},
		LabelSelector: labelSelector,
		FieldSelector: "",
	}

	return client.K8sClient.AppsV1().Deployments(namespace).Watch(ctx, opts)
}

func (client *Client) waitDeploymentReady(ctx context.Context, namespace, deployment string) error {
	watcher, err := client.createDeploymentWatcher(ctx, namespace, deployment)
	if err != nil {
		return err
	}

	defer watcher.Stop()

	wasDown := false

	for {
		select {
		case event := <-watcher.ResultChan():
			if event.Type == watch.Modified {
				if event.Object == nil {
					continue
				}

				deployment := event.Object.(*v1.Deployment)

				if deployment.Status.UnavailableReplicas > 0 {
					wasDown = true
					continue
				}

				if deployment.Status.UnavailableReplicas == 0 && deployment.Status.ReadyReplicas > 0 && wasDown {
					return nil
				}
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func (client *Client) ReadDeployment(id string) (*models.DeploymentPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read k8s deployment %s. details: %w", id, err)
	}

	if id == "" {
		log.Println("no id supplied when reading k8s deployment. assuming it was deleted")
		return nil, nil
	}

	list, err := client.K8sClient.AppsV1().Deployments(client.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, makeError(err)
	}

	for _, item := range list.Items {
		idLabel := GetLabel(item.ObjectMeta.Labels, keys.ManifestLabelID)
		if idLabel == id {
			return models.CreateDeploymentPublicFromRead(&item), nil
		}
	}

	return nil, nil
}

func (client *Client) CreateDeployment(public *models.DeploymentPublic) (*models.DeploymentPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create deployment %s. details: %w", public.Name, err)
	}

	if public.Name == "" {
		return nil, nil
	}

	deployment, err := client.K8sClient.AppsV1().Deployments(public.Namespace).Get(context.TODO(), public.Name, metav1.GetOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return nil, makeError(err)
	}

	if err == nil {
		return models.CreateDeploymentPublicFromRead(deployment), nil
	}

	public.ID = uuid.New().String()
	public.CreatedAt = time.Now()

	manifest := CreateDeploymentManifest(public)
	_, err = client.K8sClient.AppsV1().Deployments(public.Namespace).Create(context.TODO(), manifest, metav1.CreateOptions{})
	if err != nil {
		return nil, makeError(err)
	}

	return client.ReadDeployment(public.ID)
}

func (client *Client) UpdateDeployment(public *models.DeploymentPublic) (*models.DeploymentPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update k8s deployment %s. details: %w", public.Name, err)
	}

	if public.ID == "" {
		log.Println("no id supplied when updating k8s deployment. assuming it was deleted")
		return nil, nil
	}

	res, err := client.K8sClient.AppsV1().Deployments(public.Namespace).Update(context.TODO(), CreateDeploymentManifest(public), metav1.UpdateOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return nil, makeError(err)
	}

	if err == nil {
		return models.CreateDeploymentPublicFromRead(res), nil
	}

	log.Println("k8s deployment", public.Name, "not found when updating. assuming it was deleted")
	return nil, nil
}

func (client *Client) DeleteDeployment(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete deployment %s. details: %w", id, err)
	}

	if id == "" {
		log.Println("no id supplied when deleting k8s deployment. assuming it was deleted")
		return nil
	}

	list, err := client.K8sClient.AppsV1().Deployments(client.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", keys.ManifestLabelID, id),
	})
	if err != nil {
		return makeError(err)
	}

	for _, item := range list.Items {
		err = client.K8sClient.AppsV1().Deployments(client.Namespace).Delete(context.TODO(), item.Name, metav1.DeleteOptions{})
		if err != nil && !IsNotFoundErr(err) {
			return makeError(err)
		}
	}

	return nil
}

func (client *Client) RestartDeployment(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to restart k8s deployment %s. details: %w", id, err)
	}

	deployment, err := client.ReadDeployment(id)
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

	log.Println("deployment", id, "restarted")

	return nil
}

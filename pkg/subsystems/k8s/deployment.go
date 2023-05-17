package k8s

import (
	"context"
	"errors"
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

func (client *Client) ReadDeployment(namespace, id string) (*models.DeploymentPublic, error) {
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

	list, err := client.K8sClient.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, item := range list.Items {
		idLabel := GetLabel(item.ObjectMeta.Labels, keys.ManifestLabelID)
		if idLabel == id {
			return models.CreateDeploymentPublicFromRead(&item), nil
		}
	}

	return nil, nil
}

func (client *Client) CreateDeployment(public *models.DeploymentPublic) (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create deployment %s. details: %s", public.Name, err)
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

	list, err := client.K8sClient.AppsV1().Deployments(public.Namespace).List(context.TODO(), metav1.ListOptions{})
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
	manifest := CreateDeploymentManifest(public)
	_, err = client.K8sClient.AppsV1().Deployments(public.Namespace).Create(context.TODO(), manifest, metav1.CreateOptions{})
	if err != nil {
		return "", makeError(err)
	}

	return public.ID, nil
}

func (client *Client) UpdateDeployment(public *models.DeploymentPublic) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update deployment %s. details: %s", public.Name, err)
	}

	if public.ID == "" {
		return makeError(errors.New("name required"))
	}

	if public.Namespace == "" {
		return makeError(errors.New("namespace required"))
	}

	namespaceCreated, err := client.NamespaceCreated(public.Namespace)
	if err != nil {
		return makeError(err)
	}

	if !namespaceCreated {
		return makeError(fmt.Errorf("no such namespace %s", public.Namespace))
	}

	list, err := client.K8sClient.AppsV1().Deployments(public.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		idLabel := GetLabel(item.ObjectMeta.Labels, keys.ManifestLabelID)
		if idLabel == public.ID {
			manifest := CreateDeploymentManifest(public)
			_, err = client.K8sClient.AppsV1().Deployments(public.Namespace).Update(context.TODO(), manifest, metav1.UpdateOptions{})
			if err != nil {
				return makeError(err)
			}

			return nil
		}

	}

	return makeError(fmt.Errorf("no such deployment %s", public.Name))
}

func (client *Client) DeleteDeployment(namespace, id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete deployment %s. details: %s", id, err)
	}

	if id == "" {
		return makeError(errors.New("id required"))
	}

	if namespace == "" {
		return makeError(errors.New("namespace required"))
	}

	namespaceCreated, err := client.NamespaceCreated(namespace)
	if err != nil {
		return makeError(err)
	}

	if !namespaceCreated {
		return nil
	}

	list, err := client.K8sClient.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		idLabel := GetLabel(item.ObjectMeta.Labels, keys.ManifestLabelID)
		if idLabel == id {
			err = client.K8sClient.AppsV1().Deployments(namespace).Delete(context.TODO(), item.Name, metav1.DeleteOptions{})
			if err != nil {
				return makeError(err)
			}

			return nil
		}
	}

	return makeError(fmt.Errorf("no such deployment %s", id))
}

func (client *Client) RestartDeployment(public *models.DeploymentPublic) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to restart deployment %s. details: %s", public.Name, err)
	}

	namespaceCreated, err := client.NamespaceCreated(public.Namespace)
	if err != nil {
		return makeError(err)
	}

	if !namespaceCreated {
		return makeError(fmt.Errorf("no such namespace %s", public.Namespace))
	}

	req := client.K8sClient.AppsV1().Deployments(public.Namespace)

	data := fmt.Sprintf(`{"spec": {"template": {"metadata": {"annotations": {"kubectl.kubernetes.io/restartedAt": "%s"}}}}}`, time.Now().Format("20060102150405"))

	_, err = req.Patch(context.TODO(), public.Name, types.StrategicMergePatchType, []byte(data), metav1.PatchOptions{})
	if err != nil {
		return makeError(err)
	}

	err = client.waitDeploymentReady(context.TODO(), public.Namespace, public.Name)
	if err != nil {
		return makeError(err)
	}

	log.Println("deployment", public.Name, "restarted")

	return nil
}

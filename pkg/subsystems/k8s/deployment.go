package k8s

import (
	"context"
	"fmt"
	"go-deploy/pkg/subsystems/k8s/manifests"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

func (client *Client) DeploymentCreated(namespace, name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if k8s deployment %s is created. details: %s", name, err)
	}

	deployment, err := client.K8sClient.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if isNotFoundError(err) {
			return false, nil
		}
		return false, makeError(err)
	}
	return deployment != nil, err
}

func (client *Client) CreateDeployment(namespace, name string, dockerImage string, envs []apiv1.EnvVar) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s deployment %s. details: %s", name, err)
	}

	deploymentCreated, err := client.DeploymentCreated(namespace, name)
	if err != nil {
		return makeError(err)
	}

	if deploymentCreated {
		return nil
	}

	namespaceCreated, err := client.NamespaceCreated(namespace)
	if err != nil {
		return makeError(err)
	}

	if !namespaceCreated {
		err = fmt.Errorf("no namespace created")
		return makeError(err)
	}

	deployment := manifests.CreateDeploymentManifest(namespace, name, dockerImage, envs)

	_, err = client.K8sClient.AppsV1().Deployments(namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		return makeError(err)
	}

	return nil
}

func (client *Client) RestartDeployment(namespace, name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to restart k8s deployment %s. details: %s", name, err)
	}

	req := client.K8sClient.AppsV1().Deployments(namespace)

	data := fmt.Sprintf(`{"spec": {"template": {"metadata": {"annotations": {"kubectl.kubernetes.io/restartedAt": "%s"}}}}}`, time.Now().Format("20060102150405"))

	_, err := req.Patch(context.TODO(), name, types.StrategicMergePatchType, []byte(data), metav1.PatchOptions{})
	if err != nil {
		return makeError(err)
	}

	return nil
}

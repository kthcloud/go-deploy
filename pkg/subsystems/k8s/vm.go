package k8s

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/pkg/subsystems/k8s/keys"
	"go-deploy/pkg/subsystems/k8s/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func (client *Client) ReadVM(id string) (*models.VmPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read k8s vm. details: %w", err)
	}

	vm, err := client.KubeVirtK8sClient.KubevirtV1().VirtualMachines(client.Namespace).Get(context.TODO(), id, metav1.GetOptions{})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	return &models.VmPublic{
		Name: vm.Name,
	}, nil
}

func (client *Client) CreateVM(public *models.VmPublic) (*models.VmPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s vm. details: %w", err)
	}

	vms, err := client.KubeVirtK8sClient.KubevirtV1().VirtualMachines(client.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", keys.LabelDeployName, public.Name),
	})
	if err != nil && !IsNotFoundErr(err) {
		return nil, makeError(err)
	}

	if len(vms.Items) > 0 && err == nil {
		return models.CreateVmPublicFromRead(&vms.Items[0]), nil
	}

	public.ID = fmt.Sprintf("vm-%s", uuid.New().String())
	public.CreatedAt = time.Now()

	manifest := CreateVmManifest(public)
	res, err := client.KubeVirtK8sClient.KubevirtV1().VirtualMachines(client.Namespace).Create(context.TODO(), manifest, metav1.CreateOptions{})
	if err != nil {
		return nil, makeError(err)
	}

	return models.CreateVmPublicFromRead(res), nil
}

func (client *Client) UpdateVM(public *models.VmPublic) (*models.VmPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update k8s vm. details: %w", err)
	}

	manifest := CreateVmManifest(public)
	res, err := client.KubeVirtK8sClient.KubevirtV1().VirtualMachines(client.Namespace).Update(context.TODO(), manifest, metav1.UpdateOptions{})
	if err != nil {
		return nil, makeError(err)
	}

	return models.CreateVmPublicFromRead(res), nil
}

func (client *Client) DeleteVM(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s vm. details: %w", err)
	}

	err := client.KubeVirtK8sClient.KubevirtV1().VirtualMachines(client.Namespace).Delete(context.TODO(), id, metav1.DeleteOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return makeError(err)
	}

	err = client.waitVmDeleted(id)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func (client *Client) waitVmDeleted(id string) error {
	maxWait := 120
	for i := 0; i < maxWait; i++ {
		time.Sleep(1 * time.Second)
		_, err := client.KubeVirtK8sClient.KubevirtV1().VirtualMachines(client.Namespace).Get(context.TODO(), id, metav1.GetOptions{})
		if err != nil && IsNotFoundErr(err) {
			return nil
		}
	}

	return fmt.Errorf("timeout waiting for vm %s to be deleted", id)
}

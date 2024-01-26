package k8s

import (
	"context"
	"fmt"
	"go-deploy/pkg/subsystems/k8s/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func (client *Client) ReadVM(name string) (*models.VmPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read k8s vm. details: %w", err)
	}

	vm, err := client.KubeVirtK8sClient.KubevirtV1().VirtualMachines(client.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
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

	vm, err := client.KubeVirtK8sClient.KubevirtV1().VirtualMachines(client.Namespace).Get(context.TODO(), public.Name, metav1.GetOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return nil, makeError(err)
	}

	if err == nil {
		return models.CreateVmPublicFromRead(vm), nil
	}

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

func (client *Client) DeleteVM(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s vm. details: %w", err)
	}

	err := client.KubeVirtK8sClient.KubevirtV1().VirtualMachines(client.Namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return makeError(err)
	}

	return nil
}

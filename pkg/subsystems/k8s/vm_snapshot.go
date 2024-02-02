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

func (client *Client) ReadVmSnapshot(id string) (*models.VmSnapshotPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read k8s vm snapshot. details: %w", err)
	}

	vmSnapshot, err := client.KubeVirtK8sClient.SnapshotV1alpha1().VirtualMachineSnapshots(client.Namespace).Get(context.TODO(), id, metav1.GetOptions{})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	return models.CreateVmSnapshotPublicFromRead(vmSnapshot), nil
}

func (client *Client) ReadVmSnapshots(vmID string) ([]*models.VmSnapshotPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read k8s vm snapshots. details: %w", err)
	}

	vmSnapshots, err := client.KubeVirtK8sClient.SnapshotV1alpha1().VirtualMachineSnapshots(client.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", keys.LabelDeployName, vmID),
	})
	if err != nil {
		return nil, makeError(err)
	}

	vmSnapshotsPublic := make([]*models.VmSnapshotPublic, len(vmSnapshots.Items))
	for i, vmSnapshot := range vmSnapshots.Items {
		vmSnapshotsPublic[i] = models.CreateVmSnapshotPublicFromRead(&vmSnapshot)
	}

	return vmSnapshotsPublic, nil
}

func (client *Client) CreateVmSnapshot(public *models.VmSnapshotPublic) (*models.VmSnapshotPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s vm snapshot. details: %w", err)
	}

	snapshots, err := client.KubeVirtK8sClient.SnapshotV1alpha1().VirtualMachineSnapshots(client.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", keys.LabelDeployName, public.Name),
	})
	if err != nil && !IsNotFoundErr(err) {
		return nil, makeError(err)
	}

	if len(snapshots.Items) > 0 && err == nil {
		return models.CreateVmSnapshotPublicFromRead(&snapshots.Items[0]), nil
	}

	public.ID = fmt.Sprintf("vm-snapshot-%s", uuid.New().String())
	public.CreatedAt = time.Now()

	manifest := CreateVmSnapshotManifest(public)
	_, err = client.KubeVirtK8sClient.SnapshotV1alpha1().VirtualMachineSnapshots(client.Namespace).Create(context.TODO(), manifest, metav1.CreateOptions{})
	if err != nil {
		return nil, makeError(err)
	}

	// A little bit of a hack, but KubeVirt does not update the status right away,
	// and it's a bit ugly to return "Unknown" status to the user every time a
	// snapshot is created.
	time.Sleep(3 * time.Second)

	return client.ReadVmSnapshot(public.ID)
}

func (client *Client) DeleteVmSnapshot(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s vm snapshot. details: %w", err)
	}

	err := client.KubeVirtK8sClient.SnapshotV1alpha1().VirtualMachineSnapshots(client.Namespace).Delete(context.TODO(), id, metav1.DeleteOptions{})
	if err != nil && !IsNotFoundErr(err) {
		return makeError(err)
	}

	err = client.waitVmSnapshotDeleted(id)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func (client *Client) waitVmSnapshotDeleted(id string) error {
	maxWait := 120
	for i := 0; i < maxWait; i++ {
		time.Sleep(1 * time.Second)
		_, err := client.KubeVirtK8sClient.SnapshotV1alpha1().VirtualMachineSnapshots(client.Namespace).Get(context.TODO(), id, metav1.GetOptions{})
		if err != nil && IsNotFoundErr(err) {
			return nil
		}
	}

	return fmt.Errorf("timeout waiting for vm snapshot %s to be deleted", id)
}

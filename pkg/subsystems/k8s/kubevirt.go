package k8s

import (
	"context"
	"go-deploy/pkg/subsystems/k8s/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "kubevirt.io/api/core/v1"
)

func (client *Client) SetPermittedHostDevices(devices models.PermittedHostDevices) error {
	makeError := func(err error) error {
		return err
	}

	currentKubeVirt, err := client.KubeVirtK8sClient.KubevirtV1().KubeVirts("kubevirt").Get(context.TODO(), "kubevirt", metav1.GetOptions{})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil
		}

		return makeError(err)
	}

	oldDevices := currentKubeVirt.Spec.Configuration.PermittedHostDevices

	newDevices := v1.PermittedHostDevices{}
	for _, device := range devices.PciHostDevices {
		newDevices.PciHostDevices = append(newDevices.PciHostDevices, v1.PciHostDevice{
			PCIVendorSelector:        device.PciVendorSelector,
			ResourceName:             device.ResourceName,
			ExternalResourceProvider: false,
		})
	}

	// Check if any diff, assume name always has the same PciVendorSelector
	hasDiff := false
	if oldDevices == nil {
		hasDiff = true
	} else if len(oldDevices.PciHostDevices) != len(newDevices.PciHostDevices) {
		hasDiff = true
	} else {
		oldDevicesByName := make(map[string]v1.PciHostDevice)
		for _, device := range oldDevices.PciHostDevices {
			oldDevicesByName[device.ResourceName] = device
		}

		for _, device := range newDevices.PciHostDevices {
			if _, ok := oldDevicesByName[device.ResourceName]; !ok {
				hasDiff = true
				break
			}
		}
	}
	if !hasDiff {
		return nil
	}

	currentKubeVirt.Spec.Configuration.PermittedHostDevices = &newDevices

	_, err = client.KubeVirtK8sClient.KubevirtV1().KubeVirts("kubevirt").Update(context.TODO(), currentKubeVirt, metav1.UpdateOptions{})
	if err != nil {
		return makeError(err)
	}

	return nil
}

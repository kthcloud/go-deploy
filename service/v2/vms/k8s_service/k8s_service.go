package k8s_service

import (
	"context"
	"errors"
	"fmt"
	configModels "go-deploy/models/config"
	"go-deploy/models/model"
	"go-deploy/models/version"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/vm_port_repo"
	"go-deploy/pkg/db/resources/vm_repo"
	"go-deploy/pkg/log"
	kErrors "go-deploy/pkg/subsystems/k8s/errors"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service/constants"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/resources"
	"go-deploy/service/v2/vms/opts"
	"golang.org/x/exp/slices"
)

// Create sets up K8s for a VM.
func (c *Client) Create(id string, params *model.VmCreateParams) error {
	log.Println("Setting up K8s for", params.Name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup k8s for vm %s. details: %w", params.Name, err)
	}

	vm, kc, g, err := c.Get(OptsAll(id, opts.ExtraOpts{ExtraSshKeys: []string{config.Config.VM.AdminSshPublicKey}}))
	if err != nil {
		if errors.Is(err, sErrors.VmNotFoundErr) {
			log.Println("VM not found when setting up k8s for", params.Name, ". Assuming it was deleted")
			return nil
		}

		return makeError(err)
	}

	// Namespace
	namespace := g.Namespace()
	if namespace != nil {
		err = resources.SsCreator(kc.CreateNamespace).
			WithDbFunc(dbFunc(id, "namespace")).
			WithPublic(namespace).
			Exec()
		if err != nil {
			return makeError(err)
		}
	}

	// PVs
	for _, pvPublic := range g.PVs() {
		err = resources.SsCreator(kc.CreatePV).
			WithDbFunc(dbFunc(id, "pvMap."+pvPublic.Name)).
			WithPublic(&pvPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// PVCs
	for _, pvcPublic := range g.PVCs() {
		err = resources.SsCreator(kc.CreatePVC).
			WithDbFunc(dbFunc(id, "pvcMap."+pvcPublic.Name)).
			WithPublic(&pvcPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Deployment
	for _, deploymentPublic := range g.Deployments() {
		err = resources.SsCreator(kc.CreateDeployment).
			WithDbFunc(dbFunc(id, "deploymentMap."+deploymentPublic.Name)).
			WithPublic(&deploymentPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// VM
	for _, vmPublic := range g.VMs() {
		err = resources.SsCreator(kc.CreateVM).
			WithDbFunc(dbFunc(id, "vmMap."+vmPublic.Name)).
			WithPublic(&vmPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Service
	for _, servicePublic := range g.Services() {
		for idx, port := range servicePublic.Ports {
			if port.Port == 0 {
				vmPort, err := vm_port_repo.New().GetOrLeaseAny(port.TargetPort, vm.ID, vm.Zone)
				if err != nil {
					if errors.Is(err, vm_port_repo.NoPortsAvailableErr) {
						return makeError(sErrors.NoPortsAvailableErr)
					}

					return makeError(err)
				}

				servicePublic.Ports[idx].Port = vmPort.PublicPort
			}
		}

		err = resources.SsCreator(kc.CreateService).
			WithDbFunc(dbFunc(id, "serviceMap."+servicePublic.Name)).
			WithPublic(&servicePublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Ingress
	for _, ingressPublic := range g.Ingresses() {
		err = resources.SsCreator(kc.CreateIngress).
			WithDbFunc(dbFunc(id, "ingressMap."+ingressPublic.Name)).
			WithPublic(&ingressPublic).
			Exec()

		if err != nil {
			if errors.Is(err, kErrors.IngressHostInUseErr) {
				return makeError(sErrors.IngressHostInUseErr)
			}

			return makeError(err)
		}
	}

	// Secret
	for _, secretPublic := range g.Secrets() {
		err = resources.SsCreator(kc.CreateSecret).
			WithDbFunc(dbFunc(id, "secretMap."+secretPublic.Name)).
			WithPublic(&secretPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

// Delete deletes the K8s setup for a VM.
func (c *Client) Delete(id string, overwriteUserID ...string) error {
	log.Println("Deleting K8s for", id)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s for deployment %s. details: %w", id, err)
	}

	var userID string
	if len(overwriteUserID) > 0 {
		userID = overwriteUserID[0]
	}

	vm, kc, _, err := c.Get(OptsNoGenerator(id, opts.ExtraOpts{UserID: userID}))
	if err != nil {
		if errors.Is(err, sErrors.VmNotFoundErr) {
			log.Println("VM not found when deleting k8s for", id, ". Assuming it was deleted")
			return nil
		}

		return makeError(err)
	}

	// Ingress
	for mapName, ingress := range vm.Subsystems.K8s.IngressMap {
		err = resources.SsDeleter(kc.DeleteIngress).
			WithResourceID(ingress.Name).
			WithDbFunc(dbFunc(id, "ingressMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Service
	for mapName, k8sService := range vm.Subsystems.K8s.ServiceMap {
		err = resources.SsDeleter(kc.DeleteService).
			WithResourceID(k8sService.Name).
			WithDbFunc(dbFunc(id, "serviceMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Deployment
	for mapName, k8sDeployment := range vm.Subsystems.K8s.DeploymentMap {
		err = resources.SsDeleter(kc.DeleteDeployment).
			WithResourceID(k8sDeployment.Name).
			WithDbFunc(dbFunc(id, "deploymentMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// VM
	for mapName, k8sVm := range vm.Subsystems.K8s.VmMap {
		err = resources.SsDeleter(kc.DeleteVM).
			WithResourceID(k8sVm.ID).
			WithDbFunc(dbFunc(id, "vmMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Secret
	for mapName, secret := range vm.Subsystems.K8s.SecretMap {
		var deleteFunc func(id string) error
		if mapName == constants.WildcardCertSecretName {
			deleteFunc = func(string) error { return nil }
		} else {
			deleteFunc = kc.DeleteSecret
		}

		err = resources.SsDeleter(deleteFunc).
			WithResourceID(secret.Name).
			WithDbFunc(dbFunc(id, "secretMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Snapshots
	for mapName, snapshot := range vm.Subsystems.K8s.VmSnapshotMap {
		err = resources.SsDeleter(kc.DeleteVmSnapshot).
			WithResourceID(snapshot.ID).
			WithDbFunc(dbFunc(id, "vmSnapshotMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// PVCs
	for mapName, pvc := range vm.Subsystems.K8s.PvcMap {
		err = resources.SsDeleter(kc.DeletePVC).
			WithResourceID(pvc.Name).
			WithDbFunc(dbFunc(id, "pvcMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// PVs
	for mapName, pv := range vm.Subsystems.K8s.PvMap {
		err = resources.SsDeleter(kc.DeletePV).
			WithResourceID(pv.Name).
			WithDbFunc(dbFunc(id, "pvMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Namespace
	err = resources.SsDeleter(func(string) error { return nil }).
		WithResourceID(vm.Subsystems.K8s.Namespace.Name).
		WithDbFunc(dbFunc(id, "namespace")).
		Exec()

	if err != nil {
		return makeError(err)
	}

	return nil
}

// Repair repairs the K8s setup for a VM.
func (c *Client) Repair(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair k8s %s. details: %w", id, err)
	}

	vm, kc, g, err := c.Get(OptsAll(id, opts.ExtraOpts{ExtraSshKeys: []string{config.Config.VM.AdminSshPublicKey}}))
	if err != nil {
		if errors.Is(err, sErrors.VmNotFoundErr) {
			log.Println("VM not found when deleting k8s for", id, ". Assuming it was deleted")
			return nil
		}

		return makeError(err)
	}

	namespace := g.Namespace()
	if namespace != nil {
		err = resources.SsRepairer(
			kc.ReadNamespace,
			kc.CreateNamespace,
			kc.UpdateNamespace,
			func(string) error { return nil },
		).WithResourceID(namespace.Name).WithDbFunc(dbFunc(id, "namespace")).WithGenPublic(namespace).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	vms := g.VMs()
	for mapName, k8sVm := range vm.Subsystems.K8s.VmMap {
		idx := slices.IndexFunc(vms, func(v k8sModels.VmPublic) bool { return v.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(kc.DeleteVM).
				WithResourceID(k8sVm.ID).
				WithDbFunc(dbFunc(id, "vmMap."+mapName)).
				Exec()

			if err != nil {
				return makeError(err)
			}
		}
	}

	for _, public := range vms {
		err = resources.SsRepairer(
			kc.ReadVM,
			kc.CreateVM,
			kc.UpdateVM,
			func(string) error { return nil },
		).WithResourceID(public.ID).WithDbFunc(dbFunc(id, "vmMap."+public.Name)).WithGenPublic(&public).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	deployments := g.Deployments()
	for mapName, k8sDeployment := range vm.Subsystems.K8s.DeploymentMap {
		idx := slices.IndexFunc(deployments, func(d k8sModels.DeploymentPublic) bool { return d.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(kc.DeleteDeployment).
				WithResourceID(k8sDeployment.Name).
				WithDbFunc(dbFunc(id, "deploymentMap."+mapName)).
				Exec()

			if err != nil {
				return makeError(err)
			}
		}
	}
	for _, public := range deployments {
		err = resources.SsRepairer(
			kc.ReadDeployment,
			kc.CreateDeployment,
			kc.UpdateDeployment,
			kc.DeleteDeployment,
		).WithResourceID(public.Name).WithDbFunc(dbFunc(id, "deploymentMap."+public.Name)).WithGenPublic(&public).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	services := g.Services()
	for mapName, k8sService := range vm.Subsystems.K8s.ServiceMap {
		idx := slices.IndexFunc(services, func(s k8sModels.ServicePublic) bool { return s.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(kc.DeleteService).
				WithResourceID(k8sService.Name).
				WithDbFunc(dbFunc(id, "serviceMap."+mapName)).
				Exec()

			if err != nil {
				return makeError(err)
			}
		}
	}
	for _, public := range services {
		for idx, port := range public.Ports {
			if port.Port == 0 {
				vmPort, err := vm_port_repo.New().GetOrLeaseAny(port.TargetPort, vm.ID, vm.Zone)
				if err != nil {
					if errors.Is(err, vm_port_repo.NoPortsAvailableErr) {
						return makeError(sErrors.NoPortsAvailableErr)
					}

					return makeError(err)
				}

				public.Ports[idx].Port = vmPort.PublicPort
			}
		}

		err = resources.SsRepairer(
			kc.ReadService,
			kc.CreateService,
			kc.UpdateService,
			kc.DeleteService,
		).WithResourceID(public.Name).WithDbFunc(dbFunc(id, "serviceMap."+public.Name)).WithGenPublic(&public).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	ingresses := g.Ingresses()
	for mapName, ingress := range vm.Subsystems.K8s.IngressMap {
		idx := slices.IndexFunc(ingresses, func(i k8sModels.IngressPublic) bool { return i.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(kc.DeleteIngress).
				WithResourceID(ingress.Name).
				WithDbFunc(dbFunc(id, "ingressMap."+mapName)).
				Exec()

			if err != nil {
				return makeError(err)
			}
		}
	}
	for _, public := range ingresses {
		err = resources.SsRepairer(
			kc.ReadIngress,
			kc.CreateIngress,
			kc.UpdateIngress,
			kc.DeleteIngress,
		).WithResourceID(public.Name).WithDbFunc(dbFunc(id, "ingressMap."+public.Name)).WithGenPublic(&public).Exec()

		if err != nil {
			if errors.Is(err, kErrors.IngressHostInUseErr) {
				return makeError(sErrors.IngressHostInUseErr)
			}

			return makeError(err)
		}
	}

	secrets := g.Secrets()
	for mapName, secret := range vm.Subsystems.K8s.SecretMap {
		idx := slices.IndexFunc(secrets, func(s k8sModels.SecretPublic) bool { return s.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(kc.DeleteSecret).
				WithResourceID(secret.Name).
				WithDbFunc(dbFunc(id, "secretMap."+mapName)).
				Exec()

			if err != nil {
				return makeError(err)
			}
		}
	}
	for _, public := range secrets {
		err = resources.SsRepairer(
			kc.ReadSecret,
			kc.CreateSecret,
			kc.UpdateSecret,
			kc.DeleteSecret,
		).WithResourceID(public.Name).WithDbFunc(dbFunc(id, "secretMap."+public.Name)).WithGenPublic(&public).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

// AttachGPU attaches a GPU to a VM.
// If there is an existing attached GPU, it will be replaced.
func (c *Client) AttachGPU(vmID, groupName string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to attach gpu %s to vm %s. details: %w", groupName, vmID, err)
	}

	vm, kc, _, err := c.Get(OptsAll(vmID))
	if vm == nil {
		return makeError(sErrors.VmNotFoundErr)
	}

	// Set the GPU to the VM
	k8sVM := vm.Subsystems.K8s.VmMap[vm.Name]
	k8sVM.GPUs = []string{groupName}
	vm.Subsystems.K8s.VmMap[vm.Name] = k8sVM

	err = resources.SsUpdater(kc.UpdateVM).
		WithDbFunc(dbFunc(vmID, "vmMap."+vm.Name)).
		WithPublic(&k8sVM).
		Exec()
	if err != nil {
		return makeError(err)
	}

	// This requires as a restart to take effect
	err = c.DoAction(vmID, &model.VmActionParams{Action: model.ActionRestartIfRunning})
	if err != nil {
		return makeError(err)
	}

	return nil
}

// DetachGPU detaches a GPU from a VM.
func (c *Client) DetachGPU(vmID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to detach gpu from vm %s. details: %w", vmID, err)
	}

	vm, kc, _, err := c.Get(OptsAll(vmID))
	if vm == nil {
		return makeError(sErrors.VmNotFoundErr)
	}

	// Remove the GPU from the VM
	k8sVM := vm.Subsystems.K8s.VmMap[vm.Name]
	k8sVM.GPUs = []string{}
	vm.Subsystems.K8s.VmMap[vm.Name] = k8sVM

	err = resources.SsUpdater(kc.UpdateVM).
		WithDbFunc(dbFunc(vmID, "vmMap."+vm.Name)).
		WithPublic(&k8sVM).
		Exec()
	if err != nil {
		return makeError(err)
	}

	// This requires as a restart to take effect
	err = c.DoAction(vmID, &model.VmActionParams{Action: model.ActionRestartIfRunning})
	if err != nil {
		return makeError(err)
	}

	return nil
}

// DoAction performs an action on a VM.
func (c *Client) DoAction(id string, action *model.VmActionParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to perform action %s on vm %s. details: %w", action.Action, id, err)
	}

	_, kc, g, err := c.Get(OptsAll(id))
	if err != nil {
		if errors.Is(err, sErrors.VmNotFoundErr) {
			log.Println("VM not found when performing action", action.Action, "on", id, ". Assuming it was deleted")
			return nil
		}

		return makeError(err)
	}

	for _, k8sVM := range g.VMs() {
		switch action.Action {
		case model.ActionStart:
			if k8sVM.Running {
				continue
			}

			k8sVM.Running = true
			err = resources.SsUpdater(kc.UpdateVM).
				WithDbFunc(dbFunc(id, "vmMap."+k8sVM.Name)).
				WithPublic(&k8sVM).
				Exec()
			if err != nil {
				return makeError(err)
			}

		case model.ActionStop:
			if !k8sVM.Running {
				continue
			}

			k8sVM.Running = false
			err = resources.SsUpdater(kc.UpdateVM).
				WithDbFunc(dbFunc(id, "vmMap."+k8sVM.Name)).
				WithPublic(&k8sVM).
				Exec()
			if err != nil {
				return makeError(err)
			}

		case model.ActionRestart:
			// This case must be handled separately, as a Restart in KubeVirt is done by first deleting any
			// VirtualMachineInstances, and then ensuring Running is set to true.

			// 1. Delete all VirtualMachineInstances
			err = kc.DeleteVMIs(k8sVM.ID)
			if err != nil {
				return makeError(err)
			}

			// 2. Ensure Running is set to true
			if k8sVM.Running {
				continue
			}

			k8sVM.Running = true
			err = resources.SsUpdater(kc.UpdateVM).
				WithDbFunc(dbFunc(id, "vmMap."+k8sVM.Name)).
				WithPublic(&k8sVM).
				Exec()
			if err != nil {
				return makeError(err)
			}
		case model.ActionRestartIfRunning:
			if !k8sVM.Running {
				continue
			}

			return c.DoAction(id, &model.VmActionParams{Action: model.ActionRestart})
		}
	}

	return nil
}

// EnsureOwner ensures the owner of the K8s setup, by deleting and then trigger a call to Repair.
func (c *Client) EnsureOwner(id, oldOwnerID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update k8s owner for vm %s. details: %w", id, err)
	}

	// Since ownership is determined by the namespace, and the namespace owns everything,
	// we need to recreate everything

	// Delete everything in the old namespace
	// Pass in the old owner ID to use the old namespace
	err := c.Delete(id)
	if err != nil {
		return makeError(err)
	}

	// Create everything in the new namespace
	// We reset the namespace to use the VM's namespace by passing an empty string
	err = c.Repair(id)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// Synchronize synchronizes GPU groups with the backend.
// This updates the KubeVirt allowed PCI devices.
func (c *Client) Synchronize(zoneName string, gpuGroups []model.GpuGroup) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to synchronize gpu groups. details: %w", err)
	}

	// Ensure zone has KubeVirt capabilities
	zone := config.Config.GetZone(zoneName)
	if zone == nil {
		return makeError(sErrors.ZoneNotFoundErr)
	}

	if !zone.HasCapability(configModels.ZoneCapabilityVM) {
		return nil
	}

	_, kc, _, err := c.Get(OptsOnlyClient(zoneName))
	if err != nil {
		return makeError(err)
	}

	var devices k8sModels.PermittedHostDevices
	for _, gpuGroup := range gpuGroups {
		devices.PciHostDevices = append(devices.PciHostDevices, k8sModels.PciHostDevice{
			PciVendorSelector: k8sModels.CreatePciVendorSelector(gpuGroup.VendorID, gpuGroup.DeviceID),
			ResourceName:      gpuGroup.Name,
		})
	}

	err = kc.SetPermittedHostDevices(devices)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// SetupStatusWatcher sets up a status watcher for a zone.
// For every status change, it triggers the callback.
func (c *Client) SetupStatusWatcher(ctx context.Context, zone *configModels.Zone, resourceType string, callback func(string, string)) error {
	_, kc, _, err := c.Get(OptsOnlyClient(zone.Name))
	if err != nil {
		return err
	}

	return kc.SetupStatusWatcher(ctx, resourceType, callback)
}

// dbFunc returns a function that updates the K8s subsystem.
func dbFunc(id, key string) func(interface{}) error {
	return func(data interface{}) error {
		if data == nil {
			return vm_repo.New(version.V2).DeleteSubsystem(id, "k8s."+key)
		}
		return vm_repo.New(version.V2).SetSubsystem(id, "k8s."+key, data)
	}
}

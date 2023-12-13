package cs_service

import (
	"errors"
	"fmt"
	configModels "go-deploy/models/config"
	gpuModel "go-deploy/models/sys/gpu"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems"
	"go-deploy/pkg/subsystems/cs/commands"
	cErrors "go-deploy/pkg/subsystems/cs/errors"
	csModels "go-deploy/pkg/subsystems/cs/models"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/resources"
	"golang.org/x/exp/slices"
	"log"
)

func (c *Client) Create(id string, params *vmModel.CreateParams) error {
	log.Println("setting up cs for", params.Name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup cs for vm %s. details: %w", params.Name, err)
	}

	vm, csc, g, err := c.Get(OptsAll(id))
	if err != nil {
		if errors.Is(err, sErrors.VmNotFoundErr) {
			return nil
		}

		return makeError(err)
	}

	zone := config.Config.VM.GetZone(vm.Zone)
	if zone == nil {
		return makeError(sErrors.ZoneNotFoundErr)
	}

	csc.WithUserSshPublicKey(params.SshPublicKey)
	csc.WithAdminSshPublicKey(config.Config.VM.AdminSshPublicKey)

	// Service offering
	for _, soPublic := range g.SOs() {
		err = resources.SsCreator(csc.CreateServiceOffering).
			WithDbFunc(dbFunc(id, "serviceOffering")).
			WithPublic(&soPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}

		vm.Subsystems.CS.ServiceOffering = soPublic
	}

	// VM
	for _, vmPublic := range g.VMs() {
		err = resources.SsCreator(csc.CreateVM).
			WithDbFunc(dbFunc(id, "vm")).
			WithPublic(&vmPublic).
			Exec()

		if err != nil {
			_ = resources.SsDeleter(csc.DeleteServiceOffering).
				WithDbFunc(dbFunc(id, "serviceOffering")).
				Exec()

			return makeError(err)
		}

		vm.Subsystems.CS.VM = vmPublic
	}

	// Port-forwarding rules
	for _, pfrPublic := range g.PFRs() {
		if pfrPublic.PublicPort == 0 {
			pfrPublic.PublicPort, err = csc.GetFreePort(
				zone.PortRange.Start,
				zone.PortRange.End,
			)

			if err != nil {
				return makeError(err)
			}
		}

		err = resources.SsCreator(csc.CreatePortForwardingRule).
			WithDbFunc(dbFunc(id, "portForwardingRuleMap."+pfrName(&pfrPublic))).
			WithPublic(&pfrPublic).
			Exec()

		if err != nil {
			if errors.Is(err, cErrors.PortInUseErr) {
				return makeError(sErrors.PortInUseErr)
			}

			return makeError(err)
		}
	}

	return nil
}

func (c *Client) Delete(id string) error {
	log.Println("deleting cs for", id)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete cs for vm %s. details: %w", id, err)
	}

	vm, csc, _, err := c.Get(OptsNoGenerator(id))
	if err != nil {
		if errors.Is(err, sErrors.VmNotFoundErr) {
			return nil
		}

		return makeError(err)
	}

	for mapName, pfr := range vm.Subsystems.CS.GetPortForwardingRuleMap() {
		err = resources.SsDeleter(csc.DeletePortForwardingRule).
			WithResourceID(pfr.ID).
			WithDbFunc(dbFunc(id, "portForwardingRuleMap."+mapName)).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	err = resources.SsDeleter(csc.DeleteVM).
		WithResourceID(vm.Subsystems.CS.VM.ID).
		WithDbFunc(dbFunc(id, "vm")).
		Exec()

	if err != nil {
		return makeError(err)
	}

	err = resources.SsDeleter(csc.DeleteServiceOffering).
		WithResourceID(vm.Subsystems.CS.ServiceOffering.ID).
		WithDbFunc(dbFunc(id, "serviceOffering")).
		Exec()

	if err != nil {
		return makeError(err)
	}

	return nil
}

func (c *Client) Update(id string, updateParams *vmModel.UpdateParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update cs for vm %s. details: %w", id, err)
	}

	vm, csc, g, err := c.Get(OptsAll(id))
	if err != nil {
		if errors.Is(err, sErrors.VmNotFoundErr) {
			return nil
		}

		return makeError(err)
	}

	zone := config.Config.VM.GetZone(vm.Zone)
	if zone == nil {
		return makeError(sErrors.ZoneNotFoundErr)
	}

	// port-forwarding rule
	if updateParams.Ports != nil {
		pfrs := g.PFRs()

		for _, currentPfr := range vm.Subsystems.CS.GetPortForwardingRuleMap() {
			if slices.IndexFunc(pfrs, func(p csModels.PortForwardingRulePublic) bool { return pfrName(&p) == pfrName(&currentPfr) }) == -1 {
				err = resources.SsDeleter(csc.DeletePortForwardingRule).
					WithResourceID(currentPfr.ID).
					WithDbFunc(dbFunc(id, "portForwardingRuleMap."+pfrName(&currentPfr))).
					Exec()

				if err != nil {
					return makeError(err)
				}
			}
		}

		for _, pfrPublic := range pfrs {
			if _, ok := vm.Subsystems.CS.PortForwardingRuleMap[pfrName(&pfrPublic)]; !ok {
				if pfrPublic.PublicPort == 0 {
					pfrPublic.PublicPort, err = csc.GetFreePort(
						zone.PortRange.Start,
						zone.PortRange.End,
					)

					if err != nil {
						return makeError(err)
					}
				}

				err = resources.SsCreator(csc.CreatePortForwardingRule).
					WithDbFunc(dbFunc(id, "portForwardingRuleMap."+pfrName(&pfrPublic))).
					WithPublic(&pfrPublic).
					Exec()

				if err != nil {
					if errors.Is(err, cErrors.PortInUseErr) {
						return makeError(sErrors.PortInUseErr)
					}

					return makeError(err)
				}
			}
		}
	}

	// service offering
	var requiresUpdate bool
	var serviceOfferingUpdated bool
	if updateParams.CpuCores != nil {
		requiresUpdate = true
	}

	if updateParams.RAM != nil {
		requiresUpdate = true
	}

	if requiresUpdate {
		var soID *string
		if so := &vm.Subsystems.CS.ServiceOffering; subsystems.Created(so) {
			err = resources.SsDeleter(csc.DeleteServiceOffering).
				WithResourceID(so.ID).
				WithDbFunc(dbFunc(id, "serviceOffering")).
				Exec()

			if err != nil {
				return makeError(err)
			}

			_, err = c.Refresh(id)
			if err != nil {
				if errors.Is(err, sErrors.VmNotFoundErr) {
					return nil
				}

				return makeError(err)
			}

			for _, soPublic := range g.SOs() {
				err = resources.SsCreator(csc.CreateServiceOffering).
					WithDbFunc(dbFunc(id, "serviceOffering")).
					WithPublic(&soPublic).
					Exec()

				if err != nil {
					return makeError(err)
				}

				soID = &soPublic.ID
			}
		} else {
			for _, soPublic := range g.SOs() {
				err = resources.SsCreator(csc.CreateServiceOffering).
					WithDbFunc(dbFunc(id, "serviceOffering")).
					WithPublic(&soPublic).
					Exec()

				if err != nil {
					return makeError(err)
				}

				soID = &soPublic.ID
			}
		}

		// make sure the vm is using the latest service offering
		if soID != nil && vm.Subsystems.CS.VM.ServiceOfferingID != *soID {
			serviceOfferingUpdated = true

			deferFunc, err := c.stopVmIfRunning(id)
			if err != nil {
				return makeError(err)
			}

			defer deferFunc()
		}
	}

	if updateParams.Name != nil || serviceOfferingUpdated {
		_, err = c.Refresh(id)
		if err != nil {
			if errors.Is(err, sErrors.VmNotFoundErr) {
				return nil
			}

			return makeError(err)
		}

		vms := g.VMs()
		for _, vmPublic := range vms {
			err = resources.SsUpdater(csc.UpdateVM).
				WithPublic(&vmPublic).
				WithDbFunc(dbFunc(id, "vm")).
				Exec()

			if err != nil {
				return makeError(err)
			}
		}
	}

	return nil
}

func (c *Client) EnsureOwner(id, oldOwnerID string) error {
	// nothing needs to be done, but the method is kept as there is a project for networks,
	// and this could be implemented as user-specific networks are added

	return nil
}

func (c *Client) Repair(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair cs %s. details: %w", id, err)
	}

	vm, csc, g, err := c.Get(OptsAll(id))
	if err != nil {
		if errors.Is(err, sErrors.VmNotFoundErr) {
			return nil
		}

		return makeError(err)
	}

	// Service offering
	so := g.SOs()[0]
	err = resources.SsRepairer(
		csc.ReadServiceOffering,
		csc.CreateServiceOffering,
		csc.UpdateServiceOffering,
		csc.DeleteServiceOffering,
	).WithResourceID(so.ID).WithGenPublic(&so).WithDbFunc(dbFunc(id, "serviceOffering")).Exec()

	if err != nil {
		return makeError(err)
	}

	_, err = c.Refresh(id)
	if err != nil {
		if errors.Is(err, sErrors.VmNotFoundErr) {
			return nil
		}

		return makeError(err)
	}

	// VM
	csVM := g.VMs()[0]
	status, err := csc.GetVmStatus(vm.Subsystems.CS.VM.ID)
	if err != nil {
		return makeError(err)
	}

	// only repair if the vm is stopped to prevent downtime for the user
	if status == "" || status == "Stopped" {
		var gpu *gpuModel.GPU
		if gpuID := vm.GetGpuID(); gpuID != nil {
			gpu, err = gpuModel.New().GetByID(*gpuID)
			if err != nil {
				return makeError(err)
			}
		}

		if gpu != nil {
			csVM.ExtraConfig = CreateExtraConfig(gpu)
		}

		// <<NEVER>> call the "DeleteVM" method here, as it contains the persistent storage for the VM
		// (this api does not handle volume in cloudstack separately from the vm,
		// so deleting the vm will delete the persistent volume)
		err = resources.SsRepairer(
			csc.ReadVM,
			csc.CreateVM,
			csc.UpdateVM,
			func(id string) error { return nil },
		).WithResourceID(csVM.ID).WithGenPublic(&csVM).WithDbFunc(dbFunc(id, "vm")).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// Port-forwarding rules
	pfrs := g.PFRs()
	for mapName, pfr := range vm.Subsystems.CS.GetPortForwardingRuleMap() {
		idx := slices.IndexFunc(pfrs, func(p csModels.PortForwardingRulePublic) bool { return pfrName(&p) == mapName })
		if idx == -1 {
			err = resources.SsDeleter(csc.DeletePortForwardingRule).
				WithResourceID(pfr.ID).
				WithDbFunc(dbFunc(id, "portForwardingRuleMap."+pfrName(&pfr))).
				Exec()

			if err != nil {
				return makeError(err)
			}

			continue
		}
	}
	for _, pfr := range pfrs {
		err = resources.SsRepairer(
			csc.ReadPortForwardingRule,
			csc.CreatePortForwardingRule,
			csc.UpdatePortForwardingRule,
			csc.DeletePortForwardingRule,
		).WithResourceID(pfr.ID).WithDbFunc(dbFunc(id, "portForwardingRuleMap."+pfrName(&pfr))).WithGenPublic(&pfr).Exec()

		if err != nil {
			if errors.Is(err, cErrors.PortInUseErr) {
				return makeError(sErrors.PortInUseErr)
			}

			return makeError(err)
		}
	}

	return nil
}

// DoCommand executes a command on the vm
func (c *Client) DoCommand(id, csVmID string, gpuID *string, command string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to execute command %s for cs vm %s. details: %w", command, csVmID, err)
	}

	_, csc, _, err := c.Get(OptsNoGenerator(id))
	if err != nil {
		if errors.Is(err, sErrors.VmNotFoundErr) {
			return nil
		}

		return makeError(err)
	}

	var requiredHost *string
	if gpuID != nil {
		requiredHost, err = c.GetRequiredHost(*gpuID)
		if err != nil {
			return makeError(err)
		}
	}

	err = csc.DoVmCommand(csVmID, requiredHost, commands.Command(command))
	if err != nil {
		return makeError(err)
	}

	return nil
}

// CheckSuitableHost checks if the host is in the correct state to start a vm
func (c *Client) CheckSuitableHost(id, csVmID, hostName string, zone *configModels.VmZone) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if cs vm %s can be started on host %s. details: %w", csVmID, hostName, err)
	}

	_, csc, _, err := c.Get(OptsNoGenerator(id))
	if err != nil {
		if errors.Is(err, sErrors.VmNotFoundErr) {
			return nil
		}

		return makeError(err)
	}

	hasCapacity, err := csc.HasCapacity(csVmID, hostName)
	if err != nil {
		return makeError(err)
	}

	if !hasCapacity {
		return sErrors.VmTooLargeErr
	}

	err = c.CheckHostState(hostName, zone)
	if err != nil {
		if errors.Is(err, sErrors.HostNotAvailableErr) {
			return sErrors.VmTooLargeErr
		}

		return makeError(err)
	}

	return nil
}

func (c *Client) GetHostByVM(vmID string) (*csModels.HostPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get host for vm %s. details: %w", vmID, err)
	}

	vm, csc, _, err := c.Get(OptsNoGenerator(vmID))
	if err != nil {
		return nil, makeError(err)
	}

	host, err := csc.ReadHostByVM(vm.Subsystems.CS.VM.ID)
	if err != nil {
		return nil, makeError(err)
	}

	return host, nil
}

func (c *Client) GetHostByName(hostName string, zone *configModels.VmZone) (*csModels.HostPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get host %s. details: %w", hostName, err)
	}

	_, csc, _, err := c.Get(OptsOnlyClient(zone))
	if err != nil {
		return nil, makeError(err)
	}

	host, err := csc.ReadHostByName(hostName)
	if err != nil {
		return nil, makeError(err)
	}

	return host, nil
}

// CheckHostState checks if the host is in the correct state to start a vm
func (c *Client) CheckHostState(hostName string, zone *configModels.VmZone) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if host %s is in correct state. details: %w", hostName, err)
	}

	_, csc, _, err := c.Get(OptsOnlyClient(zone))
	if err != nil {
		return makeError(err)
	}

	host, err := csc.ReadHostByName(hostName)
	if err != nil {
		return makeError(err)
	}

	if host.State != "Up" || host.ResourceState != "Enabled" {
		return sErrors.HostNotAvailableErr
	}

	return nil
}

func (c *Client) GetConfiguration(zone *configModels.VmZone) (*csModels.ConfigurationPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get configuration. details: %w", err)
	}

	_, csc, _, err := c.Get(OptsOnlyClient(zone))
	if err != nil {
		return nil, makeError(err)
	}

	configuration, err := csc.ReadConfiguration()
	if err != nil {
		return nil, makeError(err)
	}

	return configuration, nil
}

func (c *Client) stopVmIfRunning(id string) (func(), error) {
	vm, csc, _, err := c.Get(OptsNoGenerator(id))
	if err != nil {
		if errors.Is(err, sErrors.VmNotFoundErr) {
			return nil, nil
		}

		return nil, err
	}

	// turn it off if it is on, but remember the status
	status, err := csc.GetVmStatus(vm.Subsystems.CS.VM.ID)
	if err != nil {
		return nil, err
	}

	if status == "Running" {
		err = csc.DoVmCommand(vm.Subsystems.CS.VM.ID, nil, commands.Stop)
		if err != nil {
			return nil, err
		}
	}

	return func() {
		// turn it on if it was on
		if status == "Running" {
			var requiredHost *string
			if gpuID := vm.GetGpuID(); gpuID != nil {
				requiredHost, err = c.GetRequiredHost(*gpuID)
				if err != nil {
					log.Println("failed to get required host for vm", vm.Name, ". details:", err)
					return
				}
			}

			err = csc.DoVmCommand(vm.Subsystems.CS.VM.ID, requiredHost, commands.Start)
			if err != nil {
				log.Println("failed to start vm", vm.Name, ". details:", err)
				return
			}
		}
	}, nil
}

func dbFunc(vmID, key string) func(interface{}) error {
	return func(data interface{}) error {
		if data == nil {
			return vmModel.New().DeleteSubsystemByID(vmID, "cs."+key)
		}
		return vmModel.New().UpdateSubsystemByID(vmID, "cs."+key, data)
	}
}

func pfrName(pfr *csModels.PortForwardingRulePublic) string {
	if pfr.Name == "__ssh" {
		return pfr.Name
	}

	return fmt.Sprintf("priv-%d-prot-%s", pfr.PrivatePort, pfr.Protocol)
}

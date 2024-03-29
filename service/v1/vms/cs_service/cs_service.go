package cs_service

import (
	"errors"
	"fmt"
	configModels "go-deploy/models/config"
	"go-deploy/models/model"
	"go-deploy/models/versions"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/gpu_repo"
	"go-deploy/pkg/db/resources/vm_port_repo"
	"go-deploy/pkg/db/resources/vm_repo"
	"go-deploy/pkg/subsystems"
	"go-deploy/pkg/subsystems/cs/commands"
	cErrors "go-deploy/pkg/subsystems/cs/errors"
	csModels "go-deploy/pkg/subsystems/cs/models"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/resources"
	"golang.org/x/exp/slices"
	"log"
	"time"
)

// Create sets up the CS setup for the VM.
//
// This include creating the VM and port-forwarding rules.
func (c *Client) Create(id string, params *model.VmCreateParams) error {
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

	// VM
	for _, vmPublic := range g.VMs() {
		err = resources.SsCreator(csc.CreateVM).
			WithDbFunc(dbFunc(id, "vm")).
			WithPublic(&vmPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}

		vm.Subsystems.CS.VM = vmPublic
	}

	time.Sleep(30 * time.Second)

	// Port-forwarding rules
	for _, pfrPublic := range g.PFRs() {
		if pfrPublic.PublicPort == 0 {
			port, err := vm_port_repo.New().GetOrLeaseAny(pfrPublic.PrivatePort, vm.ID, vm.Zone)
			if err != nil {
				if errors.Is(err, vm_port_repo.NoPortsAvailableErr) {
					return makeError(sErrors.NoPortsAvailableErr)
				}

				return makeError(err)
			}

			pfrPublic.PublicPort = port.PublicPort
		}

		err = resources.SsCreator(csc.CreatePortForwardingRule).
			WithDbFunc(dbFunc(id, "portForwardingRuleMap."+pfrName(&pfrPublic))).
			WithPublic(&pfrPublic).
			Exec()

		if err != nil {
			var portInUseErr *cErrors.PortInUseError
			if errors.As(err, &portInUseErr) {
				return makeError(sErrors.NewPortInUseErr(portInUseErr.Port))
			}

			return makeError(err)
		}
	}

	return nil
}

// Delete deletes the CS setup for the VM.
//
// This includes deleting the VM and port-forwarding rules.
// Snapshots are automatically deleted by when the VM is deleted.
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

	err = resources.SsDeleter(csc.DeleteVM).
		WithResourceID(vm.Subsystems.CS.VM.ID).
		WithDbFunc(dbFunc(id, "vm")).
		Exec()

	if err != nil {
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

	return nil
}

// Update updates a VM.
//
// It updates any of the resources associated with fields in the update params and returns an error if any.
func (c *Client) Update(id string, updateParams *model.VmUpdateParams) error {
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

	// PFR
	if updateParams.PortMap != nil {
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
					vmPort, err := vm_port_repo.New().GetOrLeaseAny(pfrPublic.PrivatePort, vm.ID, vm.Zone)
					if err != nil {
						if errors.Is(err, vm_port_repo.NoPortsAvailableErr) {
							return makeError(sErrors.NoPortsAvailableErr)
						}

						return makeError(err)
					}

					pfrPublic.PublicPort = vmPort.PublicPort
				}

				err = resources.SsCreator(csc.CreatePortForwardingRule).
					WithDbFunc(dbFunc(id, "portForwardingRuleMap."+pfrName(&pfrPublic))).
					WithPublic(&pfrPublic).
					Exec()

				if err != nil {
					var portInUseErr *cErrors.PortInUseError
					if errors.As(err, &portInUseErr) {
						return makeError(sErrors.NewPortInUseErr(portInUseErr.Port))
					}

					return makeError(err)
				}
			}
		}
	}

	hasNewSpecs := updateParams.RAM != nil || updateParams.CpuCores != nil
	if hasNewSpecs {
		deferFunc, err := c.stopVmIfRunning(id)
		if err != nil {
			return makeError(err)
		}

		defer deferFunc()
	}

	requiresUpdate := hasNewSpecs || updateParams.Name != nil
	if requiresUpdate {
		for _, vmPublic := range g.VMs() {
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

// EnsureOwner ensures the owner of the CS setup.
func (c *Client) EnsureOwner(id, oldOwnerID string) error {
	// Nothing needs to be done, but the method is kept as there is a project for networks,
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

	zone := config.Config.VM.GetZone(vm.Zone)
	if zone == nil {
		return makeError(sErrors.ZoneNotFoundErr)
	}

	csc.WithUserSshPublicKey(vm.SshPublicKey)
	csc.WithAdminSshPublicKey(config.Config.VM.AdminSshPublicKey)

	// VM
	csVM := g.VMs()[0]
	status, err := csc.GetVmStatus(vm.Subsystems.CS.VM.ID)
	if err != nil {
		return makeError(err)
	}

	// Only repair if the vm is stopped to prevent downtime for the user
	if status == "" || status == "Stopped" {
		gpu, err := gpu_repo.New().WithVM(vm.ID).Get()
		if err != nil {
			return makeError(err)
		}

		if gpu != nil {
			csVM.ExtraConfig = CreateExtraConfig(gpu)
		}

		// <<NEVER>> call the "DeleteVM" method here, as it contains the persistent storage for the VM
		// (this api does not handle volumes in cloudstack separately from the vm,
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

		vm.Subsystems.CS.VM = csVM
	}

	// Port-forwarding rules
	//// Only repair PFRs if there is a cs vm
	if subsystems.Created(&vm.Subsystems.CS.VM) {
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
			if pfr.PublicPort == 0 {
				vmPort, err := vm_port_repo.New().GetOrLeaseAny(pfr.PrivatePort, vm.ID, vm.Zone)
				if err != nil {
					if errors.Is(err, vm_port_repo.NoPortsAvailableErr) {
						return makeError(sErrors.NoPortsAvailableErr)
					}

					return makeError(err)
				}

				pfr.PublicPort = vmPort.PublicPort
			}

			err = resources.SsRepairer(
				csc.ReadPortForwardingRule,
				csc.CreatePortForwardingRule,
				func(_ *csModels.PortForwardingRulePublic) (*csModels.PortForwardingRulePublic, error) {
					return nil, nil
				},
				csc.DeletePortForwardingRule,
			).WithResourceID(pfr.ID).WithDbFunc(dbFunc(id, "portForwardingRuleMap."+pfrName(&pfr))).WithGenPublic(&pfr).Exec()

			if err != nil {
				var portInUseErr *cErrors.PortInUseError
				if errors.As(err, &portInUseErr) {
					return makeError(sErrors.NewPortInUseErr(portInUseErr.Port))
				}

				return makeError(err)
			}
		}
	}

	// Snapshot, ensure the daily, weekly and monthly snapshots are created, remove redundant required snapshots
	//// Only repair snapshots if there is a cs vm
	if subsystems.Created(&vm.Subsystems.CS.VM) {
		snapshots, err := csc.ReadAllSnapshots(vm.Subsystems.CS.VM.ID)
		if err != nil {
			return makeError(err)
		}

		snapshotMap := make(map[string]csModels.SnapshotPublic)
		required := []string{"auto-daily", "auto-weekly", "auto-monthly"}

		for _, snapshot := range snapshots {
			if snapshot.State == "Error" {
				log.Println("deleting errored snapshot", snapshot.ID, "for cs vm", snapshot.VmID)
				err = csc.DeleteSnapshot(snapshot.ID)
				if err != nil {
					return makeError(err)
				}
			}

			// We don't care about snapshots that are not in ready state or user created
			if snapshot.State != "Ready" || snapshot.UserCreated() {
				continue
			}

			if _, ok := snapshotMap[snapshot.Name]; ok {
				// Delete the older snapshot
				previous := snapshotMap[snapshot.Name]
				var deleteSnapshot csModels.SnapshotPublic
				if snapshot.CreatedAt.Before(previous.CreatedAt) {
					deleteSnapshot = snapshot
				} else {
					deleteSnapshot = previous
				}

				log.Println("deleting redundant old snapshot", deleteSnapshot.ID, "for cs vm", deleteSnapshot.VmID)
				err = csc.DeleteSnapshot(previous.ID)
				if err != nil {
					return makeError(err)
				}
			}

			snapshotMap[snapshot.Name] = snapshot
		}

		for _, name := range required {
			if _, ok := snapshotMap[name]; !ok {
				log.Println("creating missing required snapshot", name, "for vm", vm.ID)
				err = c.CreateSnapshot(id, &model.CreateSnapshotParams{
					Name:        name,
					UserCreated: false,
					Overwrite:   true,
				})
				if err != nil {
					if errors.Is(err, sErrors.BadStateErr) {
						// Automatically created snapshots could fail if a GPU is attached, so we ignore this error
						continue
					}

					return makeError(err)
				}
			}
		}
	}

	return nil
}

// DoCommand executes a command on the VM.
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

// ListAllStatus returns the status of all VMs.
func (c *Client) ListAllStatus(zone *configModels.VmZone) (map[string]string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to list all statuses. details: %w", err)
	}

	_, csc, _, err := c.Get(OptsOnlyClient(zone))
	if err != nil {
		return nil, makeError(err)
	}

	statuses, err := csc.ListAllStatus()
	if err != nil {
		return nil, makeError(err)
	}

	return statuses, nil
}

// CheckSuitableHost checks if the host is in the correct state to start a vm
func (c *Client) CheckSuitableHost(id, hostName string, zone *configModels.VmZone) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if vm %s can be started on host %s. details: %w", id, hostName, err)
	}

	vm, csc, _, err := c.Get(OptsNoGenerator(id))
	if err != nil {
		if errors.Is(err, sErrors.VmNotFoundErr) {
			return nil
		}

		return makeError(err)
	}

	hasCapacity, err := csc.HasCapacity(vm.Subsystems.CS.VM.ID, hostName)
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

// GetHostByVM retrieves the host given a VM.
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

// GetHostByName retrieves the host given its name and zone.
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

// GetConfiguration retrieves the configuration for the CS environment.
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

// stopIfRunning is a helper that stops the VM if it is running,
// and returns a function to start it again.
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
			if gpuID, err := gpu_repo.New().WithVM(vm.ID).GetID(); gpuID != nil {
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

// dbFunc returns a function that updates the CS subsystem.
func dbFunc(vmID, key string) func(interface{}) error {
	return func(data interface{}) error {
		if data == nil {
			return vm_repo.New(versions.V1).DeleteSubsystem(vmID, "cs."+key)
		}
		return vm_repo.New(versions.V1).SetSubsystem(vmID, "cs."+key, data)
	}
}

// pfrName is a helper function that returns a formatted name for a port-forwarding rule.
// It formats the specifications rather than the name, which allows safe storage in the database,
// while not restricting any user-defined names..
func pfrName(pfr *csModels.PortForwardingRulePublic) string {
	if pfr.Name == "__ssh" {
		return pfr.Name
	}

	return fmt.Sprintf("priv-%d-prot-%s", pfr.PrivatePort, pfr.Protocol)
}

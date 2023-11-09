package cs_service

import (
	"errors"
	"fmt"
	configModels "go-deploy/models/config"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/cs/commands"
	csModels "go-deploy/pkg/subsystems/cs/models"
	"go-deploy/service"
	"go-deploy/service/resources"
	"go-deploy/service/vm_service/base"
	"golang.org/x/exp/slices"
	"log"
)

func Create(vmID string, params *vmModel.CreateParams) error {
	log.Println("setting up cs for", params.Name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup cs for vm %s. details: %w", params.Name, err)
	}

	context, err := NewContext(vmID)
	if err != nil {
		if errors.Is(err, base.VmDeletedErr) {
			return nil
		}

		return makeError(err)
	}

	context.Client.WithUserSshPublicKey(params.SshPublicKey)
	context.Client.WithAdminSshPublicKey(config.Config.VM.AdminSshPublicKey)
	
	// Service offering
	for _, soPublic := range context.Generator.SOs() {
		err = resources.SsCreator(context.Client.CreateServiceOffering).
			WithDbFunc(dbFunc(vmID, "serviceOffering")).
			WithPublic(&soPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}

		context.VM.Subsystems.CS.ServiceOffering = soPublic
	}

	// VM
	for _, vmPublic := range context.Generator.VMs() {
		err = resources.SsCreator(context.Client.CreateVM).
			WithDbFunc(dbFunc(vmID, "vm")).
			WithPublic(&vmPublic).
			Exec()

		if err != nil {
			_ = resources.SsDeleter(context.Client.DeleteServiceOffering).
				WithDbFunc(dbFunc(vmID, "serviceOffering")).
				Exec()

			return makeError(err)
		}

		context.VM.Subsystems.CS.VM = vmPublic
	}

	// Port-forwarding rules
	for _, pfrPublic := range context.Generator.PFRs() {
		if pfrPublic.PublicPort == 0 {
			pfrPublic.PublicPort, err = context.Client.GetFreePort(
				context.Zone.PortRange.Start,
				context.Zone.PortRange.End,
			)

			if err != nil {
				return makeError(err)
			}
		}

		err = resources.SsCreator(context.Client.CreatePortForwardingRule).
			WithDbFunc(dbFunc(vmID, "portForwardingRuleMap."+pfrPublic.Name)).
			WithPublic(&pfrPublic).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func Delete(id string) error {
	log.Println("deleting cs for", id)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete cs for vm %s. details: %w", id, err)
	}

	context, err := NewContext(id)
	if err != nil {
		if errors.Is(err, base.VmDeletedErr) {
			return nil
		}

		return makeError(err)
	}

	for mapName, pfr := range context.VM.Subsystems.CS.GetPortForwardingRuleMap() {
		err = resources.SsDeleter(context.Client.DeletePortForwardingRule).
			WithResourceID(pfr.ID).
			WithDbFunc(dbFunc(id, "portForwardingRuleMap."+mapName)).
			Exec()
	}

	err = resources.SsDeleter(context.Client.DeleteVM).
		WithResourceID(context.VM.Subsystems.CS.VM.ID).
		WithDbFunc(dbFunc(id, "vm")).
		Exec()

	if err != nil {
		return makeError(err)
	}

	err = resources.SsDeleter(context.Client.DeleteServiceOffering).
		WithResourceID(context.VM.Subsystems.CS.ServiceOffering.ID).
		WithDbFunc(dbFunc(id, "serviceOffering")).
		Exec()

	if err != nil {
		return makeError(err)
	}

	return nil
}

func Update(vmID string, updateParams *vmModel.UpdateParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update cs for vm %s. details: %w", vmID, err)
	}

	context, err := NewContext(vmID)
	if err != nil {
		if errors.Is(err, base.VmDeletedErr) {
			return nil
		}

		return makeError(err)
	}

	// port-forwarding rule
	if updateParams.Ports != nil {
		pfrs := context.Generator.PFRs()

		for _, currentPfr := range context.VM.Subsystems.CS.GetPortForwardingRuleMap() {
			if slices.IndexFunc(pfrs, func(p csModels.PortForwardingRulePublic) bool { return p.Name == currentPfr.Name }) == -1 {
				err = resources.SsDeleter(context.Client.DeletePortForwardingRule).
					WithResourceID(currentPfr.ID).
					WithDbFunc(dbFunc(vmID, "portForwardingRuleMap."+currentPfr.Name)).
					Exec()
			}
		}

		for _, pfrPublic := range pfrs {
			if _, ok := context.VM.Subsystems.CS.PortForwardingRuleMap[pfrPublic.Name]; !ok {
				if pfrPublic.PublicPort == 0 {
					pfrPublic.PublicPort, err = context.Client.GetFreePort(
						context.Zone.PortRange.Start,
						context.Zone.PortRange.End,
					)

					if err != nil {
						return makeError(err)
					}
				}

				err = resources.SsCreator(context.Client.CreatePortForwardingRule).
					WithDbFunc(dbFunc(vmID, "portForwardingRuleMap."+pfrPublic.Name)).
					WithPublic(&pfrPublic).
					Exec()

				if err != nil {
					return makeError(err)
				}
			}
		}
	}

	// service offering
	var soID *string
	if so := &context.VM.Subsystems.CS.ServiceOffering; service.Created(so) {
		var requiresUpdate bool
		if updateParams.CpuCores != nil {
			requiresUpdate = true
		}

		if updateParams.RAM != nil {
			requiresUpdate = true
		}

		if requiresUpdate {
			err = resources.SsDeleter(context.Client.DeleteServiceOffering).
				WithResourceID(so.ID).
				WithDbFunc(dbFunc(vmID, "serviceOffering")).
				Exec()

			if err != nil {
				return makeError(err)
			}

			for _, soPublic := range context.Generator.SOs() {
				soPublic.ID = ""
				err = resources.SsCreator(context.Client.CreateServiceOffering).
					WithDbFunc(dbFunc(vmID, "serviceOffering")).
					WithPublic(&soPublic).
					Exec()

				if err != nil {
					return makeError(err)
				}

				soID = &soPublic.ID
			}
		} else {
			soID = &so.ID
		}
	} else {
		for _, soPublic := range context.Generator.SOs() {
			err = resources.SsCreator(context.Client.CreateServiceOffering).
				WithDbFunc(dbFunc(vmID, "serviceOffering")).
				WithPublic(&soPublic).
				Exec()

			if err != nil {
				return makeError(err)
			}

			soID = &soPublic.ID
		}
	}

	serviceOfferingUpdated := false

	// make sure the vm is using the latest service offering
	if soID != nil && context.VM.Subsystems.CS.VM.ServiceOfferingID != *soID {
		serviceOfferingUpdated = true

		// turn it off if it is on, but remember the status
		status, err := context.Client.GetVmStatus(context.VM.Subsystems.CS.VM.ID)
		if err != nil {
			return makeError(err)
		}

		if status == "Running" {
			err = context.Client.DoVmCommand(context.VM.Subsystems.CS.VM.ID, nil, commands.Stop)
			if err != nil {
				return makeError(err)
			}
		}

		defer func() {
			// turn it on if it was on
			if status == "Running" {
				var requiredHost *string
				if context.VM.HasGPU() {
					requiredHost, err = GetRequiredHost(context.VM.GpuID)
					if err != nil {
						log.Println("failed to get required host for vm", context.VM.Name, "in zone", context.Zone.Name, ". details:", err)
						return
					}
				}

				err = context.Client.DoVmCommand(context.VM.Subsystems.CS.VM.ID, requiredHost, commands.Start)
				if err != nil {
					log.Println("failed to start vm", context.VM.Name, "in zone", context.Zone.Name, ". details:", err)
					return
				}
			}
		}()
	}

	if updateParams.Name != nil || serviceOfferingUpdated {
		vms := context.Generator.VMs()
		for _, vmPublic := range vms {
			vmPublic.ServiceOfferingID = *soID

			err = resources.SsUpdater(context.Client.UpdateVM).
				WithPublic(&vmPublic).
				WithDbFunc(dbFunc(vmID, "vm")).
				Exec()

			if err != nil {
				return makeError(err)
			}
		}
	}

	return nil
}

func EnsureOwner(id, oldOwnerID string) error {
	// nothing needs to be done, but the method is kept as there is a project for networks,
	// and this could be implemented as user-specific networks

	return nil
}

func Repair(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair cs %s. details: %w", id, err)
	}

	context, err := NewContext(id)
	if err != nil {
		if errors.Is(err, base.VmDeletedErr) {
			return nil
		}

		return makeError(err)
	}

	// Service offering
	if service.Created(&context.VM.Subsystems.CS.ServiceOffering) {
		so := context.Generator.SOs()[0]

		err = resources.SsRepairer(
			context.Client.ReadServiceOffering,
			context.Client.CreateServiceOffering,
			context.Client.UpdateServiceOffering,
			context.Client.DeleteServiceOffering,
		).WithResourceID(so.ID).WithGenPublic(&so).WithDbFunc(dbFunc(id, "serviceOffering")).Exec()

		if err != nil {
			return makeError(err)
		}
	}

	// VM
	if service.Created(&context.VM.Subsystems.CS.VM) {
		vm := context.Generator.VMs()[0]

		err = resources.SsRepairer(
			context.Client.ReadVM,
			context.Client.CreateVM,
			context.Client.UpdateVM,
			func(id string) error { return nil },
		).WithResourceID(vm.ID).WithGenPublic(&vm).WithDbFunc(dbFunc(id, "vm")).Exec()
	}

	if err != nil {
		return makeError(err)
	}

	// Port-forwarding rules
	for mapName, pfr := range context.VM.Subsystems.CS.GetPortForwardingRuleMap() {
		pfrs := context.Generator.PFRs()
		idx := slices.IndexFunc(pfrs, func(p csModels.PortForwardingRulePublic) bool { return p.Name == mapName })
		if idx == -1 {
			err = resources.SsDeleter(context.Client.DeletePortForwardingRule).
				WithResourceID(pfr.ID).
				WithDbFunc(dbFunc(id, "portForwardingRuleMap."+pfr.Name)).
				Exec()

			if err != nil {
				return makeError(err)
			}
		}

		err = resources.SsRepairer(
			context.Client.ReadPortForwardingRule,
			context.Client.CreatePortForwardingRule,
			context.Client.UpdatePortForwardingRule,
			context.Client.DeletePortForwardingRule,
		).WithResourceID(pfrs[idx].ID).WithDbFunc(dbFunc(id, "portForwardingRuleMap."+pfrs[idx].Name)).WithGenPublic(&pfrs[idx]).Exec()
	}

	return nil
}

func DoCommand(csVmID string, gpuID *string, command, zoneName string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to execute command %s for cs vm %s. details: %w", command, csVmID, err)
	}

	context, err := NewContextWithoutVM(zoneName)
	if err != nil {
		return makeError(err)
	}

	var requiredHost *string
	if gpuID != nil {
		requiredHost, err = GetRequiredHost(*gpuID)
		if err != nil {
			return makeError(err)
		}
	}

	err = context.Client.DoVmCommand(csVmID, requiredHost, commands.Command(command))
	if err != nil {
		return makeError(err)
	}

	return nil
}

func CanStart(csVmID, hostName, zoneName string) (bool, string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if cs vm %s can be started on host %s. details: %w", csVmID, hostName, err)
	}

	context, err := NewContextWithoutVM(zoneName)
	if err != nil {
		return false, "", makeError(err)
	}

	hasCapacity, err := context.Client.HasCapacity(csVmID, hostName)
	if err != nil {
		return false, "", err
	}

	if !hasCapacity {
		return false, "Host doesn't have capacity", nil
	}

	correctState, reason, err := HostInCorrectState(hostName, context.Zone)
	if err != nil {
		return false, "", err
	}

	if !correctState {
		return false, reason, nil
	}

	return true, "", nil
}

func HostInCorrectState(hostName string, zone *configModels.VmZone) (bool, string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if host %s is in correct state. details: %w", zone.Name, err)
	}

	context, err := NewContextWithoutVM(zone.Name)
	if err != nil {
		return false, "", makeError(err)
	}

	host, err := context.Client.ReadHostByName(hostName)
	if err != nil {
		return false, "", makeError(err)
	}

	if host.State != "Up" {
		return false, "Host is not up", nil
	}

	if host.ResourceState != "Enabled" {
		return false, "Host is not enabled", nil
	}

	return true, "", nil
}

func dbFunc(vmID, key string) func(interface{}) error {
	return func(data interface{}) error {
		if data == nil {
			return vmModel.New().DeleteSubsystemByID(vmID, "cs."+key)
		}
		return vmModel.New().UpdateSubsystemByID(vmID, "cs."+key, data)
	}
}

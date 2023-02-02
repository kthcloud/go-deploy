package internal_service

import (
	"fmt"
	vmModel "go-deploy/models/vm"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/pfsense"
	psModels "go-deploy/pkg/subsystems/pfsense/models"
	"log"
	"net"
)

func CreatePfSense(name string, vmIP net.IP) (*psModels.PortForwardingRulePublic, error) {
	log.Println("setting up pfSense for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup pfSense for vm %s. details: %s", name, err)
	}

	pfSenseConf := conf.Env.PfSense

	client, err := pfsense.New(&pfsense.ClientConf{
		ApiUrl:         pfSenseConf.Url,
		Username:       pfSenseConf.Identity,
		Password:       pfSenseConf.Secret,
		PublicIP:       net.ParseIP(pfSenseConf.PublicIP),
		PortRangeStart: pfSenseConf.PortRangeStart,
		PortRangeEnd:   pfSenseConf.PortRangeEnd,
	})
	if err != nil {
		return nil, makeError(err)
	}

	vm, err := vmModel.GetByName(name)
	if err != nil {
		return nil, makeError(err)
	}

	var portForwardingRule *psModels.PortForwardingRulePublic

	if vm.Subsystems.PfSense.PortForwardingRule.ID != "" {
		// make sure this is up-to-date with the ip address requested
		if vm.Subsystems.PfSense.PortForwardingRule.LocalAddress.String() != vmIP.String() {
			err = client.DeletePortForwardingRule(vm.Subsystems.PfSense.PortForwardingRule.ID)
			if err != nil {
				return nil, makeError(err)
			}

			err = vmModel.UpdateSubsystemByName(name, "pfSense", "portForwardingRule", psModels.PortForwardingRulePublic{})
			if err != nil {
				return nil, makeError(err)
			}

			vm.Subsystems.PfSense.PortForwardingRule = psModels.PortForwardingRulePublic{}
		}
	}

	if vm.Subsystems.PfSense.PortForwardingRule.ID == "" {
		id, err := client.CreatePortForwardingRule(&psModels.PortForwardingRulePublic{
			Name:         name,
			LocalAddress: vmIP,
			LocalPort:    22,
		})
		if err != nil {
			return nil, makeError(err)
		}

		portForwardingRule, err = client.ReadPortForwardingRule(id)
		if err != nil {
			return nil, makeError(err)
		}

		err = vmModel.UpdateSubsystemByName(name, "pfSense", "portForwardingRule", *portForwardingRule)
		if err != nil {
			return nil, makeError(err)
		}
	} else {
		portForwardingRule = &vm.Subsystems.PfSense.PortForwardingRule
	}

	if err != nil {
		return nil, makeError(err)
	}

	return portForwardingRule, nil
}

func DeletePfSense(name string) error {
	log.Println("deleting pfSense for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete pfSense for vm %s. details: %s", name, err)
	}

	pfSenseConf := conf.Env.PfSense

	client, err := pfsense.New(&pfsense.ClientConf{
		ApiUrl:         pfSenseConf.Url,
		Username:       pfSenseConf.Identity,
		Password:       pfSenseConf.Secret,
		PublicIP:       net.ParseIP(pfSenseConf.PublicIP),
		PortRangeStart: pfSenseConf.PortRangeStart,
		PortRangeEnd:   pfSenseConf.PortRangeEnd,
	})
	if err != nil {
		return makeError(err)
	}

	vm, err := vmModel.GetByName(name)

	if len(vm.Subsystems.PfSense.PortForwardingRule.ID) == 0 {
		return nil
	}

	err = client.DeletePortForwardingRule(vm.Subsystems.PfSense.PortForwardingRule.ID)
	if err != nil {
		return makeError(err)
	}

	err = vmModel.UpdateSubsystemByName(name, "pfSense", "portForwardingRule", psModels.PortForwardingRulePublic{})
	if err != nil {
		return makeError(err)
	}

	return nil
}

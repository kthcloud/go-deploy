package internal_service

import (
	"fmt"
	vmModel "go-deploy/models/sys/vm"
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
		Username:       pfSenseConf.User,
		Password:       pfSenseConf.Password,
		PublicIP:       net.ParseIP(pfSenseConf.PublicIP),
		PortRangeStart: pfSenseConf.PortRange.Start,
		PortRangeEnd:   pfSenseConf.PortRange.End,
	})
	if err != nil {
		return nil, makeError(err)
	}

	vm, err := vmModel.GetByName(name)
	if err != nil {
		return nil, makeError(err)
	}

	if vm == nil {
		return nil, nil
	}

	var portForwardingRule *psModels.PortForwardingRulePublic

	ruleMap := vm.Subsystems.PfSense.PortForwardingRuleMap
	if ruleMap == nil {
		ruleMap = make(map[string]psModels.PortForwardingRulePublic)
	}

	rule, hasRule := ruleMap["ssh"]
	if hasRule && rule.ID != "" {
		// make sure this is up-to-date with the ip address requested, if not, delete it to trigger creation of a new one
		if rule.LocalAddress.String() != vmIP.String() {
			err = client.DeletePortForwardingRule(rule.ID)
			if err != nil {
				return nil, makeError(err)
			}

			delete(ruleMap, "ssh")

			err = vmModel.UpdateSubsystemByName(name, "pfSense", "portForwardingRuleMap", ruleMap)
			if err != nil {
				return nil, makeError(err)
			}
		}
	}

	rule, hasRule = ruleMap["ssh"]
	if !hasRule || rule.ID == "" {
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

		if portForwardingRule == nil {
			return nil, makeError(fmt.Errorf("failed to read port forwarding rule %s after creation", id))
		}

		ruleMap["ssh"] = *portForwardingRule

		err = vmModel.UpdateSubsystemByName(name, "pfSense", "portForwardingRuleMap", ruleMap)
		if err != nil {
			return nil, makeError(err)
		}
	} else {
		rule := vm.Subsystems.PfSense.PortForwardingRuleMap["ssh"]
		portForwardingRule = &rule
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
		Username:       pfSenseConf.User,
		Password:       pfSenseConf.Password,
		PublicIP:       net.ParseIP(pfSenseConf.PublicIP),
		PortRangeStart: pfSenseConf.PortRange.Start,
		PortRangeEnd:   pfSenseConf.PortRange.End,
	})
	if err != nil {
		return makeError(err)
	}

	vm, err := vmModel.GetByName(name)

	ruleMap := vm.Subsystems.PfSense.PortForwardingRuleMap
	rule, hasRule := ruleMap["ssh"]
	if !hasRule || rule.ID == "" {
		return nil
	}

	err = client.DeletePortForwardingRule(rule.ID)
	if err != nil {
		return makeError(err)
	}

	delete(ruleMap, "ssh")

	err = vmModel.UpdateSubsystemByName(name, "pfSense", "portForwardingRuleMap", ruleMap)
	if err != nil {
		return makeError(err)
	}

	return nil
}

package subsystems

import csModels "go-deploy/pkg/subsystems/cs/models"

func (cs *CS) GetPortForwardingRuleMap() map[string]csModels.PortForwardingRulePublic {
	if cs.PortForwardingRuleMap == nil {
		cs.PortForwardingRuleMap = make(map[string]csModels.PortForwardingRulePublic)
	}

	return cs.PortForwardingRuleMap
}

func (cs *CS) GetPortForwardingRule(name string) *csModels.PortForwardingRulePublic {
	resource, ok := cs.GetPortForwardingRuleMap()[name]
	if !ok {
		return nil
	}

	return &resource
}

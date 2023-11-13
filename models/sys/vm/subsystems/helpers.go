package subsystems

import csModels "go-deploy/pkg/subsystems/cs/models"

func (cs *CS) GetPortForwardingRuleMap() map[string]csModels.PortForwardingRulePublic {
	if cs.PortForwardingRuleMap == nil {
		cs.PortForwardingRuleMap = make(map[string]csModels.PortForwardingRulePublic)
	}

	return cs.PortForwardingRuleMap
}

func (cs *CS) GetSnapshotMap() map[string]csModels.SnapshotPublic {
	if cs.SnapshotMap == nil {
		cs.SnapshotMap = make(map[string]csModels.SnapshotPublic)
	}

	return cs.SnapshotMap
}

func (cs *CS) GetPortForwardingRule(name string) *csModels.PortForwardingRulePublic {
	resource, ok := cs.GetPortForwardingRuleMap()[name]
	if !ok {
		return nil
	}

	return &resource
}

func (cs *CS) GetSnapshotByID(id string) *csModels.SnapshotPublic {
	resource, ok := cs.GetSnapshotMap()[id]
	if !ok {
		return nil
	}

	return &resource
}

func (cs *CS) GetSnapshotByName(name string) *csModels.SnapshotPublic {
	for _, resource := range cs.GetSnapshotMap() {
		if resource.Name == name {
			return &resource
		}
	}

	return nil
}

func (cs *CS) SetSnapshot(name string, resource csModels.SnapshotPublic) {
	cs.GetSnapshotMap()[name] = resource
}

func (cs *CS) SetPortForwardingRule(name string, resource csModels.PortForwardingRulePublic) {
	cs.GetPortForwardingRuleMap()[name] = resource
}

func (cs *CS) DeleteSnapshot(name string) {
	delete(cs.GetSnapshotMap(), name)
}

func (cs *CS) DeletePortForwardingRule(name string) {
	delete(cs.GetPortForwardingRuleMap(), name)
}

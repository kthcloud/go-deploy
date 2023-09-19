package subsystems

import csModels "go-deploy/pkg/subsystems/cs/models"

type CS struct {
	ServiceOffering       csModels.ServiceOfferingPublic               `bson:"serviceOffering"`
	VM                    csModels.VmPublic                            `bson:"vm"`
	PortForwardingRuleMap map[string]csModels.PortForwardingRulePublic `bson:"portForwardingRuleMap"`
	SnapshotMap           map[string]csModels.SnapshotPublic           `bson:"snapshotMap"`
}

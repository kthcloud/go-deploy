package generators

import (
	"go-deploy/pkg/subsystems/cs/models"
)

// CsGenerator is a generator for CloudStack resources
// It is used to generate the `publics`, such as models.VmPublic and models.PortForwardingRulePublic
type CsGenerator interface {
	VMs() []models.VmPublic
	PFRs() []models.PortForwardingRulePublic
}

type CsGeneratorBase struct {
	CsGenerator
}

func (cg *CsGeneratorBase) VMs() []models.VmPublic {
	return nil
}

func (cg *CsGeneratorBase) PFRs() []models.PortForwardingRulePublic {
	return nil
}

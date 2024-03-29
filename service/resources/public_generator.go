package resources

import (
	"go-deploy/models/config"
	"go-deploy/models/model"
	"go-deploy/pkg/subsystems/k8s"
)

type Deployment struct {
	deployment *model.Deployment
	zone       *config.DeploymentZone
}

type SM struct {
	sm   *model.SM
	zone *config.DeploymentZone
}

type VM struct {
	vm             *model.VM
	vmZone         *config.VmZone
	deploymentZone *config.DeploymentZone
}

// PublicGeneratorType is a the base type of all generators
// It contains the deployment, VM, and SM that should be used to generate the publics
// Depending on the generator, some of these may be nil
type PublicGeneratorType struct {
	d Deployment
	v VM
	s SM
}

// PublicGenerator returns a new PublicGeneratorType
func PublicGenerator() *PublicGeneratorType {
	return &PublicGeneratorType{}
}

// WithDeploymentZone sets the deployment zone for the generator
func (pc *PublicGeneratorType) WithDeploymentZone(zone *config.DeploymentZone) *PublicGeneratorType {
	pc.d.zone = zone
	pc.s.zone = zone
	pc.v.deploymentZone = zone
	return pc
}

// WithVmZone sets the VM zone for the generator
func (pc *PublicGeneratorType) WithVmZone(zone *config.VmZone) *PublicGeneratorType {
	pc.v.vmZone = zone
	return pc
}

// WithDeployment sets the deployment for the generator
func (pc *PublicGeneratorType) WithDeployment(deployment *model.Deployment) *PublicGeneratorType {
	pc.d.deployment = deployment
	return pc
}

// WithSM sets the SM for the generator
func (pc *PublicGeneratorType) WithSM(sm *model.SM) *PublicGeneratorType {
	pc.s.sm = sm
	return pc
}

// WithVM sets the VM for the generator
func (pc *PublicGeneratorType) WithVM(vm *model.VM) *PublicGeneratorType {
	pc.v.vm = vm
	return pc
}

// K8s returns a new K8sGenerator
func (pc *PublicGeneratorType) K8s(client *k8s.Client) *K8sGenerator {
	return &K8sGenerator{
		PublicGeneratorType: pc,
		namespace:           client.Namespace,
		client:              client,
	}
}

// Harbor returns a new HarborGenerator
func (pc *PublicGeneratorType) Harbor(project string) *HarborGenerator {
	return &HarborGenerator{
		PublicGeneratorType: pc,
		project:             project,
	}
}

// CS returns a new CsGenerator
func (pc *PublicGeneratorType) CS() *CsGenerator {
	return &CsGenerator{
		PublicGeneratorType: pc,
	}
}

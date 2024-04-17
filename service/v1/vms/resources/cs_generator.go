package resources

import (
	configModels "go-deploy/models/config"
	"go-deploy/models/model"
	"go-deploy/pkg/subsystems"
	"go-deploy/pkg/subsystems/cs/models"
	"go-deploy/service/generators"
	"sort"
	"time"
)

type CsGenerator struct {
	generators.CsGeneratorBase

	vm   *model.VM
	zone *configModels.LegacyZone
}

func CS(vm *model.VM, zone *configModels.LegacyZone) *CsGenerator {
	return &CsGenerator{
		vm:   vm,
		zone: zone,
	}
}

// VMs returns a list of models.VmPublic that should be created
func (cg *CsGenerator) VMs() []models.VmPublic {
	var res []models.VmPublic

	if cg.vm != nil {
		csVM := models.VmPublic{
			Name:        cg.vm.Name,
			CpuCores:    cg.vm.Specs.CpuCores,
			RAM:         cg.vm.Specs.RAM,
			ExtraConfig: cg.vm.Subsystems.CS.VM.ExtraConfig,
			Tags:        createTags(cg.vm.Name, cg.vm.Name),
		}

		if v := &cg.vm.Subsystems.CS.VM; subsystems.Created(v) {
			csVM.ID = v.ID
			csVM.CreatedAt = v.CreatedAt

			for idx, tag := range csVM.Tags {
				if tag.Key == "createdAt" {
					for _, createdTag := range v.Tags {
						if createdTag.Key == "createdAt" {
							csVM.Tags[idx].Value = createdTag.Value
						}
					}
				}
			}
		}

		res = append(res, csVM)
		return res
	}

	return nil
}

// PFRs returns a list of models.PortForwardingRulePublic that should be created
func (cg *CsGenerator) PFRs() []models.PortForwardingRulePublic {
	var res []models.PortForwardingRulePublic

	if cg.vm != nil {
		portMap := cg.vm.PortMap

		for name, port := range portMap {
			res = append(res, models.PortForwardingRulePublic{
				Name:        name,
				VmID:        cg.vm.Subsystems.CS.VM.ID,
				NetworkID:   cg.zone.NetworkID,
				IpAddressID: cg.zone.IpAddressID,
				PublicPort:  0, // this is set externally
				PrivatePort: port.Port,
				Protocol:    port.Protocol,
				Tags:        createTags(port.Name, cg.vm.Name),
			})
		}

		for mapName, pfr := range cg.vm.Subsystems.CS.GetPortForwardingRuleMap() {
			for idx, port := range res {
				if port.Name == mapName {
					res[idx].ID = pfr.ID
					res[idx].CreatedAt = pfr.CreatedAt
					res[idx].PublicPort = pfr.PublicPort

					for jdx, tag := range res[idx].Tags {
						if tag.Key == "createdAt" {
							for _, createdTag := range pfr.Tags {
								if createdTag.Key == "createdAt" {
									res[idx].Tags[jdx].Value = createdTag.Value
								}
							}
						}
					}

					break
				}
			}
		}

		return res
	}

	return nil
}

// createTags is a helper function to create a list of models.Tag
func createTags(name string, deployName string) []models.Tag {
	tags := []models.Tag{
		{Key: "name", Value: name},
		{Key: "deployName", Value: deployName},
		{Key: "createdAt", Value: time.Now().Format(time.RFC3339)},
	}

	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Key < tags[j].Key
	})

	return tags
}

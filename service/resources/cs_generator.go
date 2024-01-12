package resources

import (
	"fmt"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems"
	"go-deploy/pkg/subsystems/cs/models"
	"sort"
	"time"
)

type CsGenerator struct {
	*PublicGeneratorType
}

func (cr *CsGenerator) VMs() []models.VmPublic {
	var res []models.VmPublic

	if cr.v.vm != nil {
		csVM := models.VmPublic{
			Name:        cr.v.vm.Name,
			CpuCores:    cr.v.vm.Specs.CpuCores,
			RAM:         cr.v.vm.Specs.RAM,
			TemplateID:  cr.v.vmZone.TemplateID,
			ExtraConfig: cr.v.vm.Subsystems.CS.VM.ExtraConfig,
			Tags:        createTags(cr.v.vm.Name, cr.v.vm.Name),
		}

		if v := &cr.v.vm.Subsystems.CS.VM; subsystems.Created(v) {
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

func (cr *CsGenerator) PFRs() []models.PortForwardingRulePublic {
	var res []models.PortForwardingRulePublic

	if cr.v.vm != nil {
		portMap := cr.v.vm.PortMap

		for name, port := range portMap {
			res = append(res, models.PortForwardingRulePublic{
				Name:        name,
				VmID:        cr.v.vm.Subsystems.CS.VM.ID,
				NetworkID:   cr.v.vmZone.NetworkID,
				IpAddressID: cr.v.vmZone.IpAddressID,
				PublicPort:  0, // this is set externally
				PrivatePort: port.Port,
				Protocol:    port.Protocol,
				Tags:        createTags(port.Name, cr.v.vm.Name),
			})
		}

		for mapName, pfr := range cr.v.vm.Subsystems.CS.GetPortForwardingRuleMap() {
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

func createTags(name string, deployName string) []models.Tag {
	tags := []models.Tag{
		{Key: "name", Value: name},
		{Key: "managedBy", Value: config.Config.Manager},
		{Key: "deployName", Value: deployName},
		{Key: "createdAt", Value: time.Now().Format(time.RFC3339)},
	}

	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Key < tags[j].Key
	})

	return tags
}

func pfrName(privatePort int, protocol string) string {
	return fmt.Sprintf("priv-%d-prot-%s", privatePort, protocol)
}

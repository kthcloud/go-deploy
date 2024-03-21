package model

import (
	"fmt"
	"go-deploy/dto/v1/body"
	"go-deploy/models/versions"
	"go-deploy/pkg/subsystems"
	"go-deploy/utils"
	"sort"
)

// ToDTOv1 converts a VM to a DTO.
func (vm *VM) ToDTOv1(status string, connectionString *string, teams []string, gpu *body.GpuRead, externalPortMapper map[string]int) body.VmRead {

	var vmGpu *body.VmGpuLease
	if gpu != nil && gpu.Lease != nil {
		vmGpu = &body.VmGpuLease{
			ID:           gpu.ID,
			Name:         gpu.Name,
			LeaseEnd:     gpu.Lease.End,
			LeaseExpired: gpu.Lease.Expired,
		}
	}

	ports := make([]body.PortRead, 0)
	if vm.PortMap != nil {
		for _, port := range vm.PortMap {
			if port.Name == "__ssh" {
				continue
			}

			var externalPort *int
			if externalPortMapper != nil {
				if p, ok := externalPortMapper[fmt.Sprintf("priv-%d-prot-%s", port.Port, port.Protocol)]; ok {
					pLocal := p
					externalPort = &pLocal
				}
			}

			var url *string
			var customDomainUrl *string

			if port.HttpProxy != nil {
				if ingress := vm.Subsystems.K8s.GetIngress(vm.Name + "-" + port.HttpProxy.Name); subsystems.Created(ingress) {
					if len(ingress.Hosts) > 0 {
						urlStr := "https://" + ingress.Hosts[0]
						url = &urlStr
					}
				}

				if ingress := vm.Subsystems.K8s.GetIngress(vm.Name + "-" + port.HttpProxy.Name + "-custom-domain"); subsystems.Created(ingress) {
					if len(ingress.Hosts) > 0 {
						urlStr := "https://" + ingress.Hosts[0]
						customDomainUrl = &urlStr
					}
				}
			}

			var httpProxy *body.HttpProxyRead
			if port.HttpProxy != nil {
				httpProxy = &body.HttpProxyRead{
					Name:            port.HttpProxy.Name,
					URL:             url,
					CustomDomainURL: customDomainUrl,
				}

				if port.HttpProxy.CustomDomain != nil {
					httpProxy.CustomDomain = &port.HttpProxy.CustomDomain.Domain
					httpProxy.CustomDomainSecret = &port.HttpProxy.CustomDomain.Secret
					httpProxy.CustomDomainStatus = &port.HttpProxy.CustomDomain.Status
				}
			}

			ports = append(ports, body.PortRead{
				Name:         port.Name,
				Port:         port.Port,
				ExternalPort: externalPort,
				Protocol:     port.Protocol,
				HttpProxy:    httpProxy,
			})
		}
	}

	// Sort ports by name
	sort.Slice(ports, func(i, j int) bool {
		return ports[i].Name < ports[j].Name
	})

	var host *string
	if vm.Host != nil {
		host = &vm.Host.Name
	}

	return body.VmRead{
		ID:         vm.ID,
		Name:       vm.Name,
		OwnerID:    vm.OwnerID,
		Zone:       vm.Zone,
		Host:       host,
		CreatedAt:  vm.CreatedAt,
		UpdatedAt:  utils.NonZeroOrNil(vm.UpdatedAt),
		RepairedAt: utils.NonZeroOrNil(vm.RepairedAt),
		Specs: body.Specs{
			CpuCores: vm.Specs.CpuCores,
			RAM:      vm.Specs.RAM,
			DiskSize: vm.Specs.DiskSize,
		},

		Ports:        ports,
		GPU:          vmGpu,
		SshPublicKey: vm.SshPublicKey,

		Teams: teams,

		Status:           status,
		ConnectionString: connectionString,
	}
}

// FromDTOv1 converts a VM DTO to a VM.
func (p VmCreateParams) FromDTOv1(dto *body.VmCreate, fallbackZone *string, deploymentZone *string) VmCreateParams {
	p.Name = dto.Name
	p.SshPublicKey = dto.SshPublicKey
	p.CpuCores = dto.CpuCores
	p.RAM = dto.RAM
	p.DiskSize = dto.DiskSize
	p.PortMap = make(map[string]PortCreateParams)
	p.DeploymentZone = deploymentZone
	p.Version = versions.V1

	for _, port := range dto.Ports {
		if port.Name == "__ssh" {
			continue
		}

		if port.Port == 22 {
			continue
		}

		p.PortMap[portName(port.Port, port.Protocol)] = fromPortCreateDTOv1(&port)
	}

	// Ensure there is always an SSH port
	p.PortMap["__ssh"] = PortCreateParams{
		Name:     "__ssh",
		Port:     22,
		Protocol: "tcp",
	}

	if dto.Zone != nil {
		p.Zone = *dto.Zone
	} else {
		p.Zone = *fallbackZone
	}

	return p
}

// FromDTOv1 converts a VM DTO to a VM.
func (p VmUpdateParams) FromDTOv1(dto *body.VmUpdate) VmUpdateParams {
	p.Name = dto.Name
	p.SnapshotID = dto.SnapshotID
	p.CpuCores = dto.CpuCores
	p.RAM = dto.RAM

	if dto.Ports != nil {
		portMap := make(map[string]PortUpdateParams)
		for _, port := range *dto.Ports {
			if port.Name == "__ssh" {
				continue
			}

			if port.Port == 22 {
				continue
			}

			portMap[portName(port.Port, port.Protocol)] = fromPortUpdateDTOv1(&port)
		}

		// Ensure there is always an SSH port
		portMap["__ssh"] = PortUpdateParams{
			Name:     "__ssh",
			Port:     22,
			Protocol: "tcp",
		}

		p.PortMap = &portMap
	} else {
		p.PortMap = nil
	}

	return p
}

// ToDTOv1 converts a Snapshot to a body.VmSnapshotRead.
func (sc *Snapshot) ToDTOv1() body.VmSnapshotRead {
	return body.VmSnapshotRead{
		ID:         sc.ID,
		VmID:       sc.VmID,
		Name:       sc.Name,
		ParentName: sc.ParentName,
		CreatedAt:  sc.CreatedAt,
		State:      sc.State,
		Current:    sc.Current,
	}
}

// FromDTOv1 converts a body.VmSnapshotCreate to a CreateSnapshotParams.
func (sc *CreateSnapshotParams) FromDTOv1(dto *body.VmSnapshotCreate) {
	sc.Name = dto.Name
	sc.Overwrite = false
	sc.UserCreated = true
}

// fromPortCreateDTOv1 converts a port DTO to a port.
func fromPortCreateDTOv1(port *body.PortCreate) PortCreateParams {
	var httpProxy *HttpProxyCreateParams
	if port.HttpProxy != nil {
		httpProxy = &HttpProxyCreateParams{
			Name:         port.HttpProxy.Name,
			CustomDomain: port.HttpProxy.CustomDomain,
		}
	}

	return PortCreateParams{
		Name:      port.Name,
		Port:      port.Port,
		Protocol:  port.Protocol,
		HttpProxy: httpProxy,
	}
}

// fromPortUpdateDTOv1 converts a port DTO to a port.
func fromPortUpdateDTOv1(port *body.PortUpdate) PortUpdateParams {
	var httpProxy *HttpProxyUpdateParams
	if port.HttpProxy != nil {
		httpProxy = &HttpProxyUpdateParams{
			Name:         port.HttpProxy.Name,
			CustomDomain: port.HttpProxy.CustomDomain,
		}
	}

	return PortUpdateParams{
		Name:      port.Name,
		Port:      port.Port,
		Protocol:  port.Protocol,
		HttpProxy: httpProxy,
	}
}

// portName returns the name of a port used as a key in the port map in the database.
func portName(privatePort int, protocol string) string {
	return fmt.Sprintf("priv-%d-prot-%s", privatePort, protocol)
}

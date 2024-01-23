package vm

import (
	"fmt"
	"go-deploy/models/dto/v1/body"
	"go-deploy/pkg/subsystems"
	"go-deploy/utils"
	"sort"
)

// ToDTO converts a VM to a DTO.
func (vm *VM) ToDTO(status string, connectionString *string, teams []string, gpu *body.GpuRead, externalPortMapper map[string]int) body.VmRead {

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

// FromDTO converts a VM DTO to a VM.
func (p *CreateParams) FromDTO(dto *body.VmCreate, fallbackZone *string, deploymentZone *string) {
	p.Name = dto.Name
	p.SshPublicKey = dto.SshPublicKey
	p.CpuCores = dto.CpuCores
	p.RAM = dto.RAM
	p.DiskSize = dto.DiskSize
	p.PortMap = make(map[string]PortCreateParams)
	p.DeploymentZone = deploymentZone

	for _, port := range dto.Ports {
		if port.Name == "__ssh" {
			continue
		}

		if port.Port == 22 {
			continue
		}

		p.PortMap[portName(port.Port, port.Protocol)] = fromDtoPortCreate(&port)
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
}

// FromDTO converts a VM DTO to a VM.
func (p *UpdateParams) FromDTO(dto *body.VmUpdate) {
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

			portMap[portName(port.Port, port.Protocol)] = fromDtoPortUpdate(&port)
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
}

// ToDTO converts a VM DTO to a VM.
func (sc *Snapshot) ToDTO() body.VmSnapshotRead {
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

// FromDTO converts a VM DTO to a VM.
func (sc *CreateSnapshotParams) FromDTO(dto *body.VmSnapshotCreate) {
	sc.Name = dto.Name
	sc.Overwrite = false
	sc.UserCreated = true
}

// fromDtoPortCreate converts a port DTO to a port.
func fromDtoPortCreate(port *body.PortCreate) PortCreateParams {
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

// fromDtoPortUpdate converts a port DTO to a port.
func fromDtoPortUpdate(port *body.PortUpdate) PortUpdateParams {
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

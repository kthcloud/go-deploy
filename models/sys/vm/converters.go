package vm

import (
	"go-deploy/models/dto/body"
	"go-deploy/service"
)

func (vm *VM) ToDTO(status string, connectionString *string, gpu *body.GpuRead, externalPortMapper map[string]int) body.VmRead {

	var vmGpu *body.VmGpu
	if gpu != nil && gpu.Lease != nil {
		vmGpu = &body.VmGpu{
			ID:           gpu.ID,
			Name:         gpu.Name,
			LeaseEnd:     gpu.Lease.End,
			LeaseExpired: gpu.Lease.Expired,
		}
	}

	ports := make([]body.Port, 0)
	if vm.Ports != nil && externalPortMapper != nil {
		for _, port := range vm.Ports {
			if port.Name == "__ssh" {
				continue
			}

			var externalPort *int
			if _, ok := externalPortMapper[port.Name]; ok {
				p := externalPortMapper[port.Name]
				externalPort = &p
			}

			var url *string
			var customDomainUrl *string

			if port.HttpProxy != nil {
				if ingress := vm.Subsystems.K8s.GetIngress(vm.Name + "-" + port.HttpProxy.Name); service.Created(ingress) {
					if len(ingress.Hosts) > 0 {
						urlStr := "https://" + ingress.Hosts[0]
						url = &urlStr
					}
				}

				if ingress := vm.Subsystems.K8s.GetIngress(vm.Name + "-" + port.HttpProxy.Name + "-custom-domain"); service.Created(ingress) {
					if len(ingress.Hosts) > 0 {
						urlStr := "https://" + ingress.Hosts[0]
						customDomainUrl = &urlStr
					}
				}
			}

			var httpProxy *body.VmHttpProxy
			if port.HttpProxy != nil {
				httpProxy = &body.VmHttpProxy{
					Name:            port.HttpProxy.Name,
					CustomDomain:    port.HttpProxy.CustomDomain,
					URL:             url,
					CustomDomainURL: customDomainUrl,
				}
			}

			ports = append(ports, body.Port{
				Name:         port.Name,
				Port:         port.Port,
				ExternalPort: externalPort,
				Protocol:     port.Protocol,
				HttpProxy:    httpProxy,
			})
		}
	}

	return body.VmRead{
		ID:      vm.ID,
		Name:    vm.Name,
		OwnerID: vm.OwnerID,
		Zone:    vm.Zone,
		Specs: body.Specs{
			CpuCores: vm.Specs.CpuCores,
			RAM:      vm.Specs.RAM,
			DiskSize: vm.Specs.DiskSize,
		},
		Ports:            ports,
		GPU:              vmGpu,
		SshPublicKey:     vm.SshPublicKey,
		Status:           status,
		ConnectionString: connectionString,
	}
}

func (p *CreateParams) FromDTO(dto *body.VmCreate, fallbackZone *string, deploymentZone *string) {
	p.Name = dto.Name
	p.SshPublicKey = dto.SshPublicKey
	for _, port := range dto.Ports {
		if port.Name == "__ssh" {
			continue
		}

		if port.Port == 22 {
			continue
		}

		p.Ports = append(p.Ports, fromDtoPort(&port))
	}
	p.Ports = append(p.Ports, Port{
		Name:     "__ssh",
		Port:     22,
		Protocol: "tcp",
	})

	p.CpuCores = dto.CpuCores
	p.RAM = dto.RAM
	p.DiskSize = dto.DiskSize

	if dto.Zone != nil {
		p.Zone = *dto.Zone
	} else {
		p.Zone = *fallbackZone
	}

	p.DeploymentZone = deploymentZone
}

func (p *UpdateParams) FromDTO(dto *body.VmUpdate) {
	p.SnapshotID = dto.SnapshotID
	if dto.Ports != nil {
		var ports []Port
		for _, port := range *dto.Ports {
			if port.Name == "__ssh" {
				continue
			}

			if port.Port == 22 {
				continue
			}

			ports = append(ports, fromDtoPort(&port))
		}
		ports = append(ports, Port{
			Name:     "__ssh",
			Port:     22,
			Protocol: "tcp",
		})

		p.Ports = &ports
	} else {
		p.Ports = nil
	}
	p.CpuCores = dto.CpuCores
	p.RAM = dto.RAM
}

func fromDtoPort(port *body.Port) Port {
	var httpProxy *PortHttpProxy
	if port.HttpProxy != nil {
		httpProxy = &PortHttpProxy{
			Name:         port.HttpProxy.Name,
			CustomDomain: port.HttpProxy.CustomDomain,
		}
	}

	return Port{
		Name:      port.Name,
		Port:      port.Port,
		Protocol:  port.Protocol,
		HttpProxy: httpProxy,
	}
}

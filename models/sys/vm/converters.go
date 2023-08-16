package vm

import "go-deploy/models/dto/body"

func (vm *VM) ToDTO(status string, connectionString *string, gpu *body.GpuRead, externalPortMapper map[string]int) body.VmRead {

	var vmGpu *body.VmGpu
	if gpu != nil {
		vmGpu = &body.VmGpu{
			ID:       gpu.ID,
			Name:     gpu.Name,
			LeaseEnd: gpu.Lease.End,
		}
	}

	ports := make([]body.Port, 0)
	if vm.Ports != nil && externalPortMapper != nil {
		for _, port := range vm.Ports {
			if port.Name == "__ssh" {
				continue
			}

			externalPort, ok := externalPortMapper[port.Name]
			if !ok {
				continue
			}

			ports = append(ports, body.Port{
				Name:         port.Name,
				Port:         port.Port,
				ExternalPort: externalPort,
				Protocol:     port.Protocol,
			})
		}
	}

	return body.VmRead{
		ID:               vm.ID,
		Name:             vm.Name,
		SshPublicKey:     vm.SshPublicKey,
		Ports:            ports,
		OwnerID:          vm.OwnerID,
		Status:           status,
		ConnectionString: connectionString,
		GPU:              vmGpu,
		Specs: body.Specs{
			CpuCores: vm.Specs.CpuCores,
			RAM:      vm.Specs.RAM,
			DiskSize: vm.Specs.DiskSize,
		},
	}
}

func (p *UpdateParams) FromDTO(dto *body.VmUpdate) {
	p.SnapshotID = dto.SnapshotID
	if dto.Ports != nil {
		ports := make([]Port, len(*dto.Ports))
		if dto.Ports != nil {
			for i, port := range *dto.Ports {
				ports[i] = Port{
					Name:     port.Name,
					Port:     port.Port,
					Protocol: port.Protocol,
				}
			}
		}
		p.Ports = &ports
	} else {
		p.Ports = nil
	}
	p.CpuCores = dto.CpuCores
	p.RAM = dto.RAM
}

func (p *CreateParams) FromDTO(dto *body.VmCreate) {
	p.Name = dto.Name
	p.SshPublicKey = dto.SshPublicKey
	p.Ports = make([]Port, len(dto.Ports))
	for i, port := range dto.Ports {
		p.Ports[i] = Port{
			Name:     port.Name,
			Port:     port.Port,
			Protocol: port.Protocol,
		}
	}
	p.CpuCores = dto.CpuCores
	p.RAM = dto.RAM
	p.DiskSize = dto.DiskSize
}

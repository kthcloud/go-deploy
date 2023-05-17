package vm

import "go-deploy/models/dto/body"

func (vm *VM) ToDTO(status, connectionString string, gpu *body.GpuRead) body.VmRead {

	var vmGpu *body.VmGpu
	if gpu != nil {
		vmGpu = &body.VmGpu{
			ID:       gpu.ID,
			Name:     gpu.Name,
			LeaseEnd: gpu.Lease.End,
		}
	}

	return body.VmRead{
		ID:               vm.ID,
		Name:             vm.Name,
		SshPublicKey:     vm.SshPublicKey,
		OwnerID:          vm.OwnerID,
		Status:           status,
		ConnectionString: connectionString,
		GPU:              vmGpu,
	}
}

func (p *UpdateParams) FromDTO(dto *body.VmUpdate) {
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

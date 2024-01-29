package vm

import (
	"go-deploy/models/dto/v2/body"
	gpuModels "go-deploy/models/sys/gpu"
	"go-deploy/models/versions"
	"go-deploy/utils"
	"reflect"
)

func (vm *VM) ToDTOv2(gpu *gpuModels.GPU, teams []string, sshConnectionString *string) body.VmRead {
	var host *string
	if vm.Host != nil {
		host = &vm.Host.Name
	}

	var gpuLease *body.VmGpuLease
	if gpu != nil && !reflect.DeepEqual(gpu.Lease, gpuModels.GpuLease{}) {
		gpuLease = &body.VmGpuLease{
			ID:         gpu.ID,
			Name:       gpu.Data.Name,
			LeaseEndAt: gpu.Lease.End,
			IsExpired:  gpu.Lease.IsExpired(),
		}
	}

	ports := make([]body.PortRead, 0, len(vm.PortMap))
	for _, port := range vm.PortMap {
		if port.Name == "__ssh" {
			continue
		}

		var httpProxy *body.HttpProxyRead
		if port.HttpProxy != nil {
			var customDomain *body.CustomDomainRead
			if port.HttpProxy.CustomDomain != nil {
				customDomain = &body.CustomDomainRead{
					Domain: port.HttpProxy.CustomDomain.Domain,
					Secret: port.HttpProxy.CustomDomain.Secret,
					Status: port.HttpProxy.CustomDomain.Status,
				}
			}

			httpProxy = &body.HttpProxyRead{Name: port.HttpProxy.Name, CustomDomain: customDomain}
		}

		ports = append(ports, body.PortRead{
			Name:      port.Name,
			Port:      port.Port,
			Protocol:  port.Protocol,
			HttpProxy: httpProxy,
		})
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
		Ports:               ports,
		GPU:                 gpuLease,
		SshPublicKey:        vm.SshPublicKey,
		Teams:               teams,
		Status:              vm.StatusMessage,
		SshConnectionString: sshConnectionString,
	}
}

// FromDTOv2 converts a VM DTO to a VM.
func (p CreateParams) FromDTOv2(dto *body.VmCreate, fallbackZone *string) CreateParams {
	p.Name = dto.Name
	p.SshPublicKey = dto.SshPublicKey
	p.CpuCores = dto.CpuCores
	p.RAM = dto.RAM
	p.DiskSize = dto.DiskSize
	p.PortMap = make(map[string]PortCreateParams)
	p.Version = versions.V2

	// Right now we only support one zone, since we need to make sure the cluster has KubeVirt installed
	p.Zone = *fallbackZone

	for _, port := range dto.Ports {
		if port.Name == "__ssh" {
			continue
		}

		if port.Port == 22 {
			continue
		}

		p.PortMap[portName(port.Port, port.Protocol)] = fromPortCreateDTOv2(&port)
	}

	// Ensure there is always an SSH port
	p.PortMap["__ssh"] = PortCreateParams{
		Name:     "__ssh",
		Port:     22,
		Protocol: "tcp",
	}

	return p
}

// FromDTOv2 converts a VM DTO to a VM.
func (p UpdateParams) FromDTOv2(dto *body.VmUpdate) UpdateParams {
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

			portMap[portName(port.Port, port.Protocol)] = fromPortUpdateDTOv2(&port)
		}

		// Ensure there is always an SSH port
		portMap["__ssh"] = PortUpdateParams{
			Name:     "__ssh",
			Port:     22,
			Protocol: "tcp",
		}

		p.PortMap = &portMap
	}

	return p
}

// fromPortCreateDTOv2 converts a port DTO to a port.
func fromPortCreateDTOv2(port *body.PortCreate) PortCreateParams {
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

// fromPortCreateDTOv2 converts a port DTO to a port.
func fromPortUpdateDTOv2(port *body.PortUpdate) PortUpdateParams {
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

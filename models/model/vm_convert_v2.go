package model

import (
	"go-deploy/dto/v2/body"
	"go-deploy/models/versions"
	"go-deploy/utils"
	"time"
)

// ToDTOv2 converts a VM to a body.VmRead.
func (vm *VM) ToDTOv2(gpuLease *GpuLease, teams []string, sshConnectionString *string) body.VmRead {
	var host *string
	if vm.Host != nil {
		host = &vm.Host.Name
	}

	var lease *body.VmGpuLease
	if gpuLease != nil && gpuLease.IsActive() {
		lease = &body.VmGpuLease{
			ID:         gpuLease.ID,
			Name:       gpuLease.GroupName,
			LeaseEndAt: (*gpuLease.ActivatedAt).Add(time.Duration(gpuLease.LeaseDuration) * time.Hour),
			IsExpired:  gpuLease.IsExpired(),
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
			Name:         port.Name,
			Port:         port.Port,
			ExternalPort: vm.GetExternalPort(port.Port, port.Protocol),
			Protocol:     port.Protocol,
			HttpProxy:    httpProxy,
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
		GPU:                 lease,
		SshPublicKey:        vm.SshPublicKey,
		Teams:               teams,
		Status:              vm.StatusMessage,
		SshConnectionString: sshConnectionString,
	}
}

// FromDTOv2 converts a body.VmCreate to a NotificationCreateParams.
func (p VmCreateParams) FromDTOv2(dto *body.VmCreate, fallbackZone *string) VmCreateParams {
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

// FromDTOv2 converts a body.VmUpdate to a UpdateParams.
func (p VmUpdateParams) FromDTOv2(dto *body.VmUpdate) VmUpdateParams {
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

// FromDTOv2 converts a dto.VmAction to a ActionParams.
func (p VmActionParams) FromDTOv2(dto *body.VmAction) VmActionParams {
	p.Action = dto.Action
	return p
}

// ToDTOv2 converts a Snapshot to a body.VmSnapshotRead.
func (sc *SnapshotV2) ToDTOv2() body.VmSnapshotRead {
	return body.VmSnapshotRead{
		ID:        sc.ID,
		Name:      sc.Name,
		Status:    sc.Status,
		CreatedAt: sc.CreatedAt,
	}
}

// FromDTOv2 converts a body.VmSnapshotCreate to a CreateSnapshotParams.
func (sc *CreateSnapshotParams) FromDTOv2(dto *body.VmSnapshotCreate) {
	sc.Name = dto.Name
	sc.Overwrite = false
	sc.UserCreated = true
}

// fromPortCreateDTOv2 converts a body.PortCreate to a PortCreateParams.
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

// fromPortCreateDTOv2 converts a body.PortCreate to a PortCreateParams.
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

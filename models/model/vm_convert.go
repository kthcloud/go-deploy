package model

import (
	"fmt"
	"go-deploy/dto/v2/body"
	"go-deploy/pkg/subsystems"
	"go-deploy/utils"
	"time"
)

// ToDTOv2 converts a VM to a body.VmRead.
func (vm *VM) ToDTOv2(gpuLease *GpuLease, teams []string, externalPort *int, sshConnectionString *string) body.VmRead {
	var host *string
	if vm.Host != nil {
		host = &vm.Host.Name
	}

	var lease *body.VmGpuLease
	if gpuLease != nil && gpuLease.IsActive() {
		var expiresAt *time.Time
		if gpuLease.IsActive() {
			expiresAt = utils.NonZeroOrNil((*gpuLease.ActivatedAt).Add(time.Duration(gpuLease.LeaseDuration) * time.Hour))
		}

		lease = &body.VmGpuLease{
			ID:            gpuLease.ID,
			GpuGroupID:    gpuLease.GpuGroupID,
			LeaseDuration: gpuLease.LeaseDuration,
			ActivatedAt:   gpuLease.ActivatedAt,
			AssignedAt:    gpuLease.AssignedAt,
			CreatedAt:     gpuLease.CreatedAt,
			ExpiresAt:     expiresAt,
			ExpiredAt:     gpuLease.ExpiredAt,
		}
	}

	ports := make([]body.PortRead, 0, len(vm.PortMap))
	for _, port := range vm.PortMap {
		if port.Name == "__ssh" {
			continue
		}

		var httpProxy *body.HttpProxyRead
		if port.HttpProxy != nil {
			extPortStr := ""
			if externalPort != nil && *externalPort != 443 {
				extPortStr = fmt.Sprintf(":%d", *externalPort)
			}

			var customDomain *body.CustomDomainRead
			if port.HttpProxy.CustomDomain != nil {
				customDomain = &body.CustomDomainRead{
					Domain: port.HttpProxy.CustomDomain.Domain,
					URL:    fmt.Sprintf("https://%s%s", port.HttpProxy.CustomDomain.Domain, extPortStr),
					Secret: port.HttpProxy.CustomDomain.Secret,
					Status: port.HttpProxy.CustomDomain.Status,
				}
			}

			httpProxy = &body.HttpProxyRead{
				Name:         port.HttpProxy.Name,
				URL:          vm.GetHttpProxyURL(port.HttpProxy.Name, externalPort),
				CustomDomain: customDomain,
			}
		}

		ports = append(ports, body.PortRead{
			Name:         port.Name,
			Port:         port.Port,
			ExternalPort: vm.GetExternalPort(port.Port, port.Protocol),
			Protocol:     port.Protocol,
			HttpProxy:    httpProxy,
		})
	}

	var internalName *string
	if k8sVM := vm.Subsystems.K8s.VM; subsystems.Created(&k8sVM) {
		internalName = &k8sVM.ID
	}

	return body.VmRead{
		ID:           vm.ID,
		Name:         vm.Name,
		InternalName: internalName,
		OwnerID:      vm.OwnerID,
		Zone:         vm.Zone,
		Host:         host,

		CreatedAt:  vm.CreatedAt,
		UpdatedAt:  utils.NonZeroOrNil(vm.UpdatedAt),
		RepairedAt: utils.NonZeroOrNil(vm.RepairedAt),
		AccessedAt: vm.AccessedAt,

		Specs: body.VmSpecs{
			CpuCores: vm.Specs.CpuCores,
			RAM:      vm.Specs.RAM,
			DiskSize: vm.Specs.DiskSize,
		},
		Ports:               ports,
		GPU:                 lease,
		SshPublicKey:        vm.SshPublicKey,
		Teams:               teams,
		Status:              vm.Status,
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

	if dto.Zone == nil {
		p.Zone = *fallbackZone
	} else {
		p.Zone = *dto.Zone
	}

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

// FromDTOv2 converts a dto.Action to a VmActionParams.
func (p VmActionParams) FromDTOv2(dto *body.VmActionCreate) VmActionParams {
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

// portName returns the name of a port used as a key in the port map in the database.
func portName(privatePort int, protocol string) string {
	return fmt.Sprintf("priv-%d-prot-%s", privatePort, protocol)
}

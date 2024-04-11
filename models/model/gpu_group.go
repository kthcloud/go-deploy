package model

import (
	"fmt"
	"go-deploy/dto/v2/body"
	"go-deploy/utils"
)

// GpuGroup represents a group of GPUs for VM v2
type GpuGroup struct {
	// ID is just a hash of the name and zone to avoid any special characters in the name
	// To generate it use:
	//
	// utils.HashStringAlphanumericLower(fmt.Sprintf("%s-%s", Name, Zone)
	ID string `bson:"id"`
	// Name is the unique name for a group of GPUs.
	// This is used when attaching GPUs to a VM to create a host-agnostic identifier
	//
	// The name should be RFC1035 compliant, and is normally vendor/model, such as "nvidia/tesla-t4"
	Name        string `bson:"name"`
	DisplayName string `bson:"displayName"`
	Zone        string `bson:"zone"`
	Total       int    `bson:"total"`

	Vendor   string `bson:"vendor"`
	VendorID string `bson:"vendorId"`
	DeviceID string `bson:"deviceId"`
}

// ToDTO converts a model.GpuGroup to a body.GpuGroupRead DTO.
func (gpuGroup *GpuGroup) ToDTO(leases int) body.GpuGroupRead {
	// ID is just a hash of the name to avoid any special characters in the name
	id := utils.HashStringAlphanumericLower(fmt.Sprintf("%s-%s", gpuGroup.Name, gpuGroup.Zone))

	return body.GpuGroupRead{
		ID:          id,
		Name:        gpuGroup.Name,
		DisplayName: gpuGroup.DisplayName,
		Zone:        gpuGroup.Zone,
		Vendor:      gpuGroup.Vendor,
		Total:       gpuGroup.Total,
		Available:   gpuGroup.Total - leases,
	}
}

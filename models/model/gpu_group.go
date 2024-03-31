package model

import "go-deploy/dto/v2/body"

// GpuGroup represents a group of GPUs for VM v2
type GpuGroup struct {
	// ID is the unique name for a group of GPUs.
	// This is used when attaching GPUs to a VM to create a host-agnostic identifier
	//
	// The name should be RFC1035 compliant is normally vendor/model, for instance "nvidia/tesla-t4"
	ID    string `bson:"id"`
	Zone  string `bson:"zone"`
	Total int    `bson:"total"`

	Vendor   string `bson:"vendor"`
	VendorID string `bson:"vendorId"`
	DeviceID string `bson:"deviceId"`
}

// ToDTO converts a model.GpuGroup to a body.GpuGroupRead DTO.
func (gpuGroup *GpuGroup) ToDTO(leases int) body.GpuGroupRead {
	return body.GpuGroupRead{
		Name:      gpuGroup.ID,
		Total:     gpuGroup.Total,
		Available: gpuGroup.Total - leases,
	}
}

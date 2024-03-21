package model

import "time"

type GpuData struct {
	Name     string `bson:"name"`
	Slot     string `bson:"slot"`
	Vendor   string `bson:"vendor"`
	VendorID string `bson:"vendorId"`
	Bus      string `bson:"bus"`
	DeviceID string `bson:"deviceId"`
}

// Lease represents a lease of a GPU.
// Deprecated: Use gpu_repo.GPU instead
type Lease struct {
	VmID   string    `bson:"vmId"`
	UserID string    `bson:"user"`
	End    time.Time `bson:"end"`
}

func (gpuLease *Lease) IsExpired() bool {
	return gpuLease.End.Before(time.Now())
}

type GPU struct {
	ID   string `bson:"id"`
	Name string `bson:"name"`
	// GroupName is the name of the group that the GPU belongs to
	// This is used when attaching GPUs to a VM to create a host-agnostic identifier
	//
	// The name should be RFC1035 compliant is normally vendor/model, for instance "nvidia/tesla-t4"
	GroupName string `bson:"groupName"`

	Host string `bson:"host"`
	Zone string `bson:"zone"`

	Slot     string `bson:"slot"`
	Vendor   string `bson:"vendor"`
	VendorID string `bson:"vendorId"`
	Bus      string `bson:"bus"`
	DeviceID string `bson:"deviceId"`

	// Data
	// Deprecated: Use fields in GPU instead
	Data GpuData `bson:"data"`

	// Lease
	// Deprecated: Use gpu_lease.GpuLease instead
	Lease Lease `bson:"lease"`
}

func (gpu *GPU) IsAttached() bool {
	return gpu.Lease.VmID != ""
}

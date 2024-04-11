package model

import "time"

// GPU is a representation of a GPU for VM v1
type GPU struct {
	ID   string `bson:"id"`
	Name string `bson:"name"`

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

type GpuData struct {
	Name     string `bson:"name"`
	Slot     string `bson:"slot"`
	Vendor   string `bson:"vendor"`
	VendorID string `bson:"vendorId"`
	Bus      string `bson:"bus"`
	DeviceID string `bson:"deviceId"`
}

// Lease represents a lease of a GPU for VM v1
type Lease struct {
	VmID   string    `bson:"vmId"`
	UserID string    `bson:"user"`
	End    time.Time `bson:"end"`
}

func (gpuLease *Lease) IsExpired() bool {
	return gpuLease.End.Before(time.Now())
}

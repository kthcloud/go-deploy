package gpu

import "time"

type GpuData struct {
	Name     string `bson:"name" json:"name"`
	Slot     string `bson:"slot" json:"slot"`
	Vendor   string `bson:"vendor" json:"vendor"`
	VendorID string `bson:"vendorId" json:"vendorId"`
	Bus      string `bson:"bus" json:"bus"`
	DeviceID string `bson:"deviceId" json:"deviceId"`
}

type GpuLease struct {
	VmID   string    `bson:"vmId" json:"vmId"`
	UserID string    `bson:"user" json:"userId"`
	End    time.Time `bson:"end" json:"end"`
}

func (gpuLease *GpuLease) IsExpired() bool {
	return gpuLease.End.Before(time.Now())
}

type GPU struct {
	ID    string   `bson:"id" json:"id"`
	Host  string   `bson:"host" json:"host"`
	Lease GpuLease `bson:"lease" json:"lease"`
	Data  GpuData  `bson:"data" json:"data"`
	Zone  string   `bson:"zone" json:"zone"`
}

func (gpu *GPU) IsAttached() bool {
	return gpu.Lease.VmID != ""
}

package gpu

import "time"

type GpuData struct {
	Name     string `bson:"name"`
	Slot     string `bson:"slot"`
	Vendor   string `bson:"vendor"`
	VendorID string `bson:"vendorId"`
	Bus      string `bson:"bus"`
	DeviceID string `bson:"deviceId"`
}

type GpuLease struct {
	VmID   string    `bson:"vmId"`
	UserID string    `bson:"user"`
	End    time.Time `bson:"end"`
}

func (gpuLease *GpuLease) IsExpired() bool {
	return gpuLease.End.Before(time.Now())
}

type GPU struct {
	ID    string   `bson:"id"`
	Host  string   `bson:"host"`
	Lease GpuLease `bson:"lease"`
	Data  GpuData  `bson:"data"`
	Zone  string   `bson:"zone"`
}

func (gpu *GPU) IsAttached() bool {
	return gpu.Lease.VmID != ""
}

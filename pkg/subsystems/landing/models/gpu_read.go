package models

import "time"

type GpuRead struct {
	GpuInfo struct {
		Hosts []struct {
			Name string `json:"name"`
			GPUs []struct {
				Name     string `json:"name"`
				Slot     string `json:"slot"`
				Vendor   string `json:"vendor"`
				VendorID string `json:"vendorId"`
				Bus      string `json:"bus"`
				DeviceID string `json:"deviceId"`
			} `json:"gpus"`
		} `json:"hosts"`
	} `json:"gpuInfo"`
	Timestamp time.Time `json:"timestamp"`
}

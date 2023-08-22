package models

import "time"

type GpuInfoRead struct {
	GpuInfo struct {
		Hosts []struct {
			Name   string `json:"name"`
			ZoneID string `json:"zoneId"`
			GPUs   []struct {
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
	ZoneID    string    `json:"zoneId"`
}

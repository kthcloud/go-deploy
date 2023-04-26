package models

type GPU struct {
	Name string `json:"name"`
	BusID string `json:"busID"`

}

type CapabilityPublic struct {
	HostName string `json:"name"`
	GpuName string `json:"gpuName"`
}

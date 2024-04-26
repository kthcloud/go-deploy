package model

type Host struct {
	Name string `bson:"name"`
}

type PortHttpProxy struct {
	Name         string        `bson:"name,omitempty"`
	CustomDomain *CustomDomain `bson:"customDomain,omitempty"`
}

type Port struct {
	Name      string         `bson:"name"`
	Port      int            `bson:"port"`
	Protocol  string         `bson:"protocol"`
	HttpProxy *PortHttpProxy `bson:"httpProxy,omitempty"`
}

type Subsystems struct {
	CS  VmCS  `bson:"cs"`
	K8s VmK8s `bson:"k8s"`
}

type VmUsage struct {
	CpuCores  int `bson:"cpuCores"`
	RAM       int `bson:"ram"`
	DiskSize  int `bson:"diskSize"`
	Snapshots int `bson:"snapshots"`
}

type VmTransfer struct {
	Code   string `bson:"code"`
	UserID string `bson:"userId"`
}

type CloudStackHostCapabilities struct {
	CpuCoresTotal int
	CpuCoresUsed  int
	RamTotal      int
	RamUsed       int
	RamAllocated  int
}

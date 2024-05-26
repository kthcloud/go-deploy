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
	K8s VmK8s `bson:"k8s"`
}

type VmUsage struct {
	CpuCores int `bson:"cpuCores"`
	RAM      int `bson:"ram"`
	DiskSize int `bson:"diskSize"`
}

type VmStatus struct {
	Name            string `bson:"name"`
	PrintableStatus string `bson:"printableStatus"`
}

type VmiStatus struct {
	Name string  `bson:"name"`
	Host *string `bson:"host,omitempty"`
}

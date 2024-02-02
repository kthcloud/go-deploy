package vm

import (
	"go-deploy/models/sys/vm/subsystems"
)

const (
	// CustomDomainStatusPending is the status of a custom domain that is pending verification.
	CustomDomainStatusPending = "pending"
	// CustomDomainStatusVerificationFailed is the status of a custom domain that failed verification.
	// This is either caused by the DNS record not being set or the DNS record not being propagated yet.
	CustomDomainStatusVerificationFailed = "verificationFailed"
	// CustomDomainStatusActive is the status of a custom domain that is active,
	// i.e., the DNS record is set and propagated.
	CustomDomainStatusActive = "active"
)

type Host struct {
	ID   string `bson:"id"`
	Name string `bson:"name"`
}

type CustomDomain struct {
	Domain string `bson:"domain"`
	Secret string `bson:"secret"`
	Status string `bson:"status"`
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
	CS  subsystems.CS  `bson:"cs"`
	K8s subsystems.K8s `bson:"k8s"`
}

type Usage struct {
	CpuCores  int `bson:"cpuCores"`
	RAM       int `bson:"ram"`
	DiskSize  int `bson:"diskSize"`
	Snapshots int `bson:"snapshots"`
}

type Transfer struct {
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

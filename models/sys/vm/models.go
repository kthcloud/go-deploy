package vm

import (
	"go-deploy/models/sys/vm/subsystems"
)

const (
	CustomDomainStatusPending            = "pending"
	CustomDomainStatusVerificationFailed = "verificationFailed"
	CustomDomainStatusActive             = "active"
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

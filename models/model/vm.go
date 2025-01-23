package model

import (
	"fmt"
	"time"
)

type VM struct {
	ID      string `bson:"id"`
	Name    string `bson:"name"`
	Version string `bson:"version"`
	Zone    string `bson:"zone"`
	OwnerID string `bson:"ownerId"`

	CreatedAt  time.Time `bson:"createdAt"`
	UpdatedAt  time.Time `bson:"updatedAt,omitempty"`
	RepairedAt time.Time `bson:"repairedAt,omitempty"`
	DeletedAt  time.Time `bson:"deletedAt,omitempty"`
	AccessedAt time.Time `bson:"accessedAt"`

	NeverStale bool `bson:"neverStale"`

	SshPublicKey string          `bson:"sshPublicKey"`
	PortMap      map[string]Port `bson:"portMap"`
	Specs        VmSpecs         `bson:"specs"`

	Subsystems Subsystems          `bson:"subsystems"`
	Activities map[string]Activity `bson:"activities"`

	// Host is the host where the VM is running
	// It is set by the status updater worker
	Host *VmHost `bson:"host,omitempty"`
	// Status is the current status of a VM instance
	// It is set by the status updater worker
	Status string `bson:"status"`
}

type VmSpecs struct {
	CpuCores int `json:"cpuCores"`
	RAM      int `json:"ram"`
	DiskSize int `json:"diskSize"`
}

func (vm *VM) Ready() bool {
	return !vm.DoingActivity(ActivityBeingCreated) && !vm.DoingActivity(ActivityBeingDeleted)
}

func (vm *VM) DoingActivity(activity string) bool {
	for _, a := range vm.Activities {
		if a.Name == activity {
			return true
		}
	}
	return false
}

func (vm *VM) DoingOneOfActivities(activities []string) bool {
	for _, a := range activities {
		if vm.DoingActivity(a) {
			return true
		}
	}
	return false
}

func (vm *VM) BeingCreated() bool {
	return vm.DoingActivity(ActivityBeingCreated)
}

func (vm *VM) BeingDeleted() bool {
	return vm.DoingActivity(ActivityBeingDeleted)
}

// GetHttpProxyURL returns the URL of a VM's HTTP proxy.
// If the K8s ingress does not exist, it will return nil, or if the ingress does not have a host, it will return nil.
func (vm *VM) GetHttpProxyURL(name string, externalPort *int) *string {
	ingress := vm.Subsystems.K8s.GetIngress(fmt.Sprintf("%s-%s", vm.Name, name))
	if ingress == nil || !ingress.Created() {
		return nil
	}

	if len(ingress.Hosts) > 0 && len(ingress.Hosts[0]) > 0 {
		url := fmt.Sprintf("https://%s", ingress.Hosts[0])

		// If we have a custom port, we need to append it to the URL
		if externalPort != nil && *externalPort != 443 {
			url = fmt.Sprintf("%s:%d", url, *externalPort)
		}

		return &url
	}

	return nil
}

func (vm *VM) GetExternalPort(privatePort int, protocol string) *int {
	pfrName := fmt.Sprintf("priv-%d-prot-%s", privatePort, protocol)
	service := vm.Subsystems.K8s.GetService(fmt.Sprintf("%s-%s", vm.Name, pfrName))
	if service == nil {
		return nil
	}

	for _, port := range service.Ports {
		if port.Name == pfrName {
			return &port.Port
		}
	}

	return nil
}

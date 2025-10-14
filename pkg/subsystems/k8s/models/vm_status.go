package models

import (
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/keys"
	kubevirtv1 "kubevirt.io/api/core/v1"
)

type VmStatus struct {
	Name            string
	PrintableStatus string
}

type VmiStatus struct {
	Name string
	Host *string
}

func CreateVmStatusFromRead(vm *kubevirtv1.VirtualMachine) *VmStatus {
	var deployName string
	if n, ok := vm.Labels[keys.LabelDeployName]; ok {
		deployName = n
	}

	return &VmStatus{
		Name:            deployName,
		PrintableStatus: string(vm.Status.PrintableStatus),
	}
}

func CreateVmiStatusFromRead(vmi *kubevirtv1.VirtualMachineInstance) *VmiStatus {
	var deployName string
	if n, ok := vmi.Labels[keys.LabelDeployName]; ok {
		deployName = n
	}

	var host *string
	if h, ok := vmi.Labels["kubevirt.io/nodeName"]; ok {
		host = &h
	}

	return &VmiStatus{
		Name: deployName,
		Host: host,
	}
}

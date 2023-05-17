package models

import "go-deploy/pkg/imp/cloudstack"

type ServiceOfferingPublic struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`

	CpuCores int `json:"cpuCores"`
	RAM      int `json:"ram"`
}

func CreateServiceOfferingPublicFromGet(serviceOffering *cloudstack.ServiceOffering) *ServiceOfferingPublic {
	return &ServiceOfferingPublic{
		ID:          serviceOffering.Id,
		Name:        serviceOffering.Name,
		Description: serviceOffering.Displaytext,
		CpuCores:    serviceOffering.Cpunumber,
		RAM:         serviceOffering.Memory / 1024,
	}
}

package models

import (
	"go-deploy/pkg/imp/cloudstack"
	"time"
)

type ServiceOfferingPublic struct {
	ID          string    `bson:"id"`
	Name        string    `bson:"name"`
	Description string    `bson:"description"`
	CpuCores    int       `bson:"cpuCores"`
	RAM         int       `bson:"ram"`
	DiskSize    int       `bson:"diskSize"`
	CreatedAt   time.Time `bson:"createdAt"`
}

func (serviceOffering *ServiceOfferingPublic) Created() bool {
	return serviceOffering.ID != ""
}

func CreateServiceOfferingPublicFromGet(serviceOffering *cloudstack.ServiceOffering) *ServiceOfferingPublic {
	return &ServiceOfferingPublic{
		ID:          serviceOffering.Id,
		Name:        serviceOffering.Name,
		Description: serviceOffering.Displaytext,
		CpuCores:    serviceOffering.Cpunumber,
		RAM:         serviceOffering.Memory / 1024,
		DiskSize:    int(serviceOffering.Rootdisksize),
		CreatedAt:   formatCreatedAt(serviceOffering.Created),
	}
}

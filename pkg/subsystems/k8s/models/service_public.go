package models

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
)

type ServicePublic struct {
	ID         string `bson:"id"`
	Name       string `bson:"name"`
	Namespace  string `bson:"namespace"`
	Port       int    `bson:"port"`
	TargetPort int    `bson:"targetPort"`
}

func (service *ServicePublic) GetHostName() string {
	hostName := fmt.Sprintf("%s-%s", service.Name, service.ID)
	return hostName
}

func CreateServicePublicFromRead(service *v1.Service) *ServicePublic {
	idAndName, err := GetIdAndName(service.Name)
	if err != nil {
		panic(err)
	}

	return &ServicePublic{
		ID:         idAndName[0],
		Name:       idAndName[1],
		Namespace:  service.Namespace,
		Port:       int(service.Spec.Ports[0].Port),
		TargetPort: service.Spec.Ports[0].TargetPort.IntValue(),
	}
}

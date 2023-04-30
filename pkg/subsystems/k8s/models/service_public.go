package models

import (
	"go-deploy/pkg/subsystems/k8s/keys"
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
	return service.Name
}

func CreateServicePublicFromRead(service *v1.Service) *ServicePublic {
	return &ServicePublic{
		ID:         service.Labels[keys.ManifestLabelID],
		Name:       service.Labels[keys.ManifestLabelName],
		Namespace:  service.Namespace,
		Port:       int(service.Spec.Ports[0].Port),
		TargetPort: service.Spec.Ports[0].TargetPort.IntValue(),
	}
}

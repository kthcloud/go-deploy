package models

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"time"
)

type ServicePublic struct {
	Name       string            `bson:"name"`
	Namespace  string            `bson:"namespace"`
	Port       int               `bson:"port"`
	TargetPort int               `bson:"targetPort"`
	ExternalIP *string           `bson:"externalIP"`
	Selector   map[string]string `bson:"selector"`
	CreatedAt  time.Time         `bson:"createdAt"`
}

func (s *ServicePublic) Created() bool {
	return s.CreatedAt != time.Time{}
}

func (s *ServicePublic) IsPlaceholder() bool {
	return false
}

func (s *ServicePublic) GetFQDN() string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", s.Name, s.Namespace)
}

// CreateServicePublicFromRead creates a ServicePublic from a v1.Service.
func CreateServicePublicFromRead(service *v1.Service) *ServicePublic {
	var selector map[string]string
	if service.Spec.Selector != nil {
		selector = service.Spec.Selector
	}

	var externalIP *string
	if len(service.Spec.ExternalIPs) > 0 {
		externalIP = &service.Spec.ExternalIPs[0]
	}

	return &ServicePublic{
		Name:       service.Name,
		Namespace:  service.Namespace,
		Port:       int(service.Spec.Ports[0].Port),
		TargetPort: service.Spec.Ports[0].TargetPort.IntValue(),
		CreatedAt:  formatCreatedAt(service.Annotations),
		Selector:   selector,
		ExternalIP: externalIP,
	}
}

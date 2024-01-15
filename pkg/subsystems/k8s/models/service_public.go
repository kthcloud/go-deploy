package models

import (
	"fmt"
	"go-deploy/pkg/subsystems/k8s/keys"
	v1 "k8s.io/api/core/v1"
	"time"
)

type ServicePublic struct {
	ID         string    `bson:"id"`
	Name       string    `bson:"name"`
	Namespace  string    `bson:"namespace"`
	Port       int       `bson:"port"`
	TargetPort int       `bson:"targetPort"`
	CreatedAt  time.Time `bson:"createdAt"`
}

func (s *ServicePublic) Created() bool {
	return s.ID != ""
}

func (s *ServicePublic) IsPlaceholder() bool {
	return false
}

func (s *ServicePublic) GetFQDN() string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", s.Name, s.Namespace)
}

// CreateServicePublicFromRead creates a ServicePublic from a v1.Service.
func CreateServicePublicFromRead(service *v1.Service) *ServicePublic {
	return &ServicePublic{
		ID:         service.Labels[keys.ManifestLabelID],
		Name:       service.Name,
		Namespace:  service.Namespace,
		Port:       int(service.Spec.Ports[0].Port),
		TargetPort: service.Spec.Ports[0].TargetPort.IntValue(),
		CreatedAt:  formatCreatedAt(service.Annotations),
	}
}

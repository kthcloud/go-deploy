package models

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"strings"
	"time"
)

type Port struct {
	Name       string `bson:"name"`
	Protocol   string `bson:"protocol"`
	Port       int    `bson:"port"`
	TargetPort int    `bson:"targetPort"`
}

type ServicePublic struct {
	Name           string            `bson:"name"`
	Namespace      string            `bson:"namespace"`
	Ports          []Port            `bson:"ports"`
	LoadBalancerIP *string           `bson:"loadBalancerIp"`
	Selector       map[string]string `bson:"selector"`
	CreatedAt      time.Time         `bson:"createdAt"`
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

func (s *ServicePublic) IsNodePort() bool {
	for _, port := range s.Ports {
		if port.Port > 30000 && port.Port < 32767 {
			return true
		}
	}

	return false
}

// CreateServicePublicFromRead creates a ServicePublic from a v1.Service.
func CreateServicePublicFromRead(service *v1.Service) *ServicePublic {
	var selector map[string]string
	if service.Spec.Selector != nil {
		selector = service.Spec.Selector
	}

	var loadBalancerIP *string
	if len(service.Status.LoadBalancer.Ingress) > 0 {
		loadBalancerIP = &service.Status.LoadBalancer.Ingress[0].IP
	}

	var ports []Port
	for _, port := range service.Spec.Ports {
		ports = append(ports, Port{
			Name:       port.Name,
			Protocol:   strings.ToLower(string(port.Protocol)),
			Port:       int(port.Port),
			TargetPort: port.TargetPort.IntValue(),
		})
	}

	return &ServicePublic{
		Name:           service.Name,
		Namespace:      service.Namespace,
		Ports:          ports,
		CreatedAt:      formatCreatedAt(service.Annotations),
		Selector:       selector,
		LoadBalancerIP: loadBalancerIP,
	}
}

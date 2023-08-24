package models

import (
	"go-deploy/pkg/subsystems/k8s/keys"
	v1 "k8s.io/api/networking/v1"
)

type IngressPublic struct {
	ID           string   `json:"id" bson:"id"`
	Name         string   `bson:"name"`
	Namespace    string   `bson:"namespace"`
	ServiceName  string   `bson:"serviceName"`
	ServicePort  int      `bson:"servicePort"`
	IngressClass string   `bson:"ingressClassName"`
	Hosts        []string `bson:"host"`
	Placeholder  bool     `bson:"placeholder"`
}

func (i *IngressPublic) Created() bool {
	return i.ID != ""
}

func (i *IngressPublic) IsPlaceholder() bool {
	return i.Placeholder
}

func CreateIngressPublicFromRead(ingress *v1.Ingress) *IngressPublic {
	var serviceName string
	var servicePort int
	var ingressClassName string

	if len(ingress.Spec.Rules) > 0 {
		if len(ingress.Spec.Rules[0].HTTP.Paths) > 0 {
			serviceName = ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name
			servicePort = int(ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Port.Number)
		}
	}

	if ingress.Annotations != nil {
		ingressClassName = ingress.Annotations[keys.K8sAnnotationIngressClass]
	}

	hosts := make([]string, 0)
	for _, rule := range ingress.Spec.Rules {
		hosts = append(hosts, rule.Host)
	}

	return &IngressPublic{
		ID:           ingress.Labels[keys.ManifestLabelID],
		Name:         ingress.Labels[keys.ManifestLabelName],
		Namespace:    ingress.Namespace,
		ServiceName:  serviceName,
		ServicePort:  servicePort,
		IngressClass: ingressClassName,
		Hosts:        hosts,
		Placeholder:  false,
	}
}

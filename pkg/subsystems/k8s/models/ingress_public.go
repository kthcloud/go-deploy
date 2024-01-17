package models

import (
	"go-deploy/pkg/subsystems/k8s/keys"
	v1 "k8s.io/api/networking/v1"
	"time"
)

type CustomCert struct {
	ClusterIssuer string `bson:"clusterIssuer"`
	CommonName    string `bson:"commonName"`
}

type IngressPublic struct {
	Name         string      `bson:"name"`
	Namespace    string      `bson:"namespace"`
	ServiceName  string      `bson:"serviceName"`
	ServicePort  int         `bson:"servicePort"`
	IngressClass string      `bson:"ingressClassName"`
	Hosts        []string    `bson:"host"`
	Placeholder  bool        `bson:"placeholder"`
	CreatedAt    time.Time   `bson:"createdAt"`
	CustomCert   *CustomCert `bson:"customCert,omitempty"`
	TlsSecret    *string     `bson:"tlsSecret,omitempty"`
}

func (i *IngressPublic) Created() bool {
	return i.CreatedAt != time.Time{}
}

func (i *IngressPublic) IsPlaceholder() bool {
	return i.Placeholder
}

// CreateIngressPublicFromRead creates a IngressPublic from a v1.Ingress.
func CreateIngressPublicFromRead(ingress *v1.Ingress) *IngressPublic {
	var serviceName string
	var servicePort int
	var ingressClassName string
	var customCert *CustomCert

	if ingress.Spec.IngressClassName != nil {
		ingressClassName = *ingress.Spec.IngressClassName
	}

	if len(ingress.Spec.Rules) > 0 {
		if len(ingress.Spec.Rules[0].HTTP.Paths) > 0 {
			serviceName = ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name
			servicePort = int(ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Port.Number)
		}
	}

	if ingress.Annotations != nil {
		clusterIssuer := ingress.Annotations[keys.K8sAnnotationClusterIssuer]
		commonName := ingress.Annotations[keys.K8sAnnotationCommonName]

		if clusterIssuer != "" && commonName != "" {
			customCert = &CustomCert{
				ClusterIssuer: clusterIssuer,
				CommonName:    commonName,
			}
		}
	}

	hosts := make([]string, 0)
	for _, rule := range ingress.Spec.Rules {
		hosts = append(hosts, rule.Host)
	}

	var tlsSecret *string
	if customCert == nil && len(ingress.Spec.TLS) > 0 {
		tlsSecret = &ingress.Spec.TLS[0].SecretName
	}

	return &IngressPublic{
		Name:         ingress.Name,
		Namespace:    ingress.Namespace,
		ServiceName:  serviceName,
		ServicePort:  servicePort,
		IngressClass: ingressClassName,
		Hosts:        hosts,
		Placeholder:  false,
		CreatedAt:    formatCreatedAt(ingress.Annotations),
		CustomCert:   customCert,
		TlsSecret:    tlsSecret,
	}
}

package models

import (
	v1 "k8s.io/api/networking/v1"
	"time"
)

type EgressRule struct {
	// CIDR defines what is allowed by the network policy
	CIDR string `bson:"allowedCidr"`
	// Except are the blocked CIDRs, and are subsets of CIDR
	Except []string `bson:"except"`
}

type IngressRule struct {
	// CIDR defines what is allowed by the network policy
	CIDR string `bson:"allowedCidr"`
	// Except are the blocked CIDRs, and are subsets of CIDR
	Except []string `bson:"except"`
}

type NetworkPolicyPublic struct {
	Name         string            `bson:"name"`
	Namespace    string            `bson:"namespace"`
	EgressRules  []EgressRule      `bson:"egress,omitempty"`
	IngressRules []IngressRule     `bson:"ingress,omitempty"`
	Selector     map[string]string `bson:"selector,omitempty"`
	CreatedAt    time.Time         `bson:"createdAt"`
}

func (npo *NetworkPolicyPublic) Created() bool {
	return !npo.CreatedAt.IsZero()
}

func (npo *NetworkPolicyPublic) IsPlaceholder() bool {
	return false
}

func CreateNetworkPolicyPublicFromRead(policy *v1.NetworkPolicy) *NetworkPolicyPublic {
	var egressRules []EgressRule
	if len(policy.Spec.Egress) > 0 {
		for _, e := range policy.Spec.Egress[0].To {
			egress := EgressRule{
				CIDR:   e.IPBlock.CIDR,
				Except: e.IPBlock.Except,
			}
			egressRules = append(egressRules, egress)
		}
	}

	var ingressRules []IngressRule
	if len(policy.Spec.Ingress) > 0 {
		for _, i := range policy.Spec.Ingress[0].From {
			ingress := IngressRule{
				CIDR:   i.IPBlock.CIDR,
				Except: i.IPBlock.Except,
			}
			ingressRules = append(ingressRules, ingress)
		}
	}

	return &NetworkPolicyPublic{
		Name:         policy.Name,
		Namespace:    policy.Namespace,
		EgressRules:  egressRules,
		IngressRules: ingressRules,
		Selector:     policy.Spec.PodSelector.MatchLabels,
		CreatedAt:    formatCreatedAt(policy.Annotations),
	}
}

package models

import (
	v1 "k8s.io/api/networking/v1"
	"time"
)

type IpBlock struct {
	// CIDR defines what is allowed by the network policy
	CIDR string `bson:"cidr"`
	// Except are the blocked CIDRs, and are subsets of CIDR
	Except []string `bson:"except"`
}

type EgressRule struct {
	// IpBlock defines what is allowed by the network policy
	IpBlock *IpBlock `bson:"ipBlock,omitempty"`
	// PodSelector defines what is allowed by the network policy
	PodSelector map[string]string `bson:"podSelector,omitempty"`
	// NamespaceSelector defines what is allowed by the network policy
	NamespaceSelector map[string]string `bson:"namespaceSelector,omitempty"`
}

type IngressRule struct {
	// IpBlock defines what is allowed by the network policy
	IpBlock *IpBlock `bson:"ipBlock,omitempty"`
	// PodSelector defines what is allowed by the network policy
	PodSelector map[string]string `bson:"podSelector,omitempty"`
	// NamespaceSelector defines what is allowed by the network policy
	NamespaceSelector map[string]string `bson:"namespaceSelector,omitempty"`
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
			var egress EgressRule

			if e.IPBlock != nil {
				egress.IpBlock = &IpBlock{
					CIDR:   e.IPBlock.CIDR,
					Except: e.IPBlock.Except,
				}
			}

			if e.PodSelector != nil {
				egress.PodSelector = e.PodSelector.MatchLabels
			}

			if e.NamespaceSelector != nil {
				egress.NamespaceSelector = e.NamespaceSelector.MatchLabels
			}

			egressRules = append(egressRules, egress)
		}
	}

	var ingressRules []IngressRule
	if len(policy.Spec.Ingress) > 0 {
		for _, i := range policy.Spec.Ingress[0].From {
			var ingress IngressRule

			if i.IPBlock != nil {
				ingress.IpBlock = &IpBlock{
					CIDR:   i.IPBlock.CIDR,
					Except: i.IPBlock.Except,
				}
			}

			if i.PodSelector != nil {
				ingress.PodSelector = i.PodSelector.MatchLabels
			}

			if i.NamespaceSelector != nil {
				ingress.NamespaceSelector = i.NamespaceSelector.MatchLabels
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

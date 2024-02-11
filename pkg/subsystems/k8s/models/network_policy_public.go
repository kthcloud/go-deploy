package models

import (
	v1 "k8s.io/api/networking/v1"
	"time"
)

type EgressRule struct {
	// CIDR defines what is allowed by the network policy
	CIDR string `bson:"allowedCidr"`
	// Except are the blocked CIDRs, and are subsets of CIDR
	Except []string `bson:"exceptCidrs"`
}

type NetworkPolicyPublic struct {
	Name        string       `bson:"name"`
	Namespace   string       `bson:"namespace"`
	EgressRules []EgressRule `bson:"egress,omitempty"`
	CreatedAt   time.Time    `bson:"createdAt"`
}

func (npo *NetworkPolicyPublic) Created() bool {
	return !npo.CreatedAt.IsZero()
}

func (npo *NetworkPolicyPublic) IsPlaceholder() bool {
	return false
}

func CreateNetworkPolicyPublicFromRead(policy *v1.NetworkPolicy) *NetworkPolicyPublic {
	var egressRules []EgressRule

	if policy.Spec.Egress != nil && len(policy.Spec.Egress) > 0 {
		for _, e := range policy.Spec.Egress[0].To {
			egress := EgressRule{
				CIDR:   e.IPBlock.CIDR,
				Except: e.IPBlock.Except,
			}
			egressRules = append(egressRules, egress)
		}
	}

	return &NetworkPolicyPublic{
		Name:        policy.Name,
		Namespace:   policy.Namespace,
		EgressRules: egressRules,
		CreatedAt:   formatCreatedAt(policy.Annotations),
	}
}

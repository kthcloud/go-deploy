package models

import (
	v1 "k8s.io/api/core/v1"
	"time"
)

type NamespacePublic struct {
	Name      string    `bson:"name"`
	CreatedAt time.Time `bson:"createdAt"`
}

func (n *NamespacePublic) Created() bool {
	return n.CreatedAt != time.Time{}
}

func (n *NamespacePublic) IsPlaceholder() bool {
	return false
}

// CreateNamespacePublicFromRead creates a NamespacePublic from a v1.Namespace.
func CreateNamespacePublicFromRead(namespace *v1.Namespace) *NamespacePublic {
	return &NamespacePublic{
		Name:      namespace.Name,
		CreatedAt: formatCreatedAt(namespace.Annotations),
	}
}

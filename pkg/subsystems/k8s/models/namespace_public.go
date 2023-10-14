package models

import (
	"go-deploy/pkg/subsystems/k8s/keys"
	v1 "k8s.io/api/core/v1"
	"time"
)

type NamespacePublic struct {
	ID        string    `bson:"id"`
	Name      string    `bson:"name"`
	CreatedAt time.Time `bson:"createdAt"`
}

func (n *NamespacePublic) Created() bool {
	return n.ID != ""
}

func (n *NamespacePublic) IsPlaceholder() bool {
	return false
}

func CreateNamespacePublicFromRead(namespace *v1.Namespace) *NamespacePublic {
	return &NamespacePublic{
		ID:        namespace.Labels[keys.ManifestLabelID],
		Name:      namespace.Name,
		CreatedAt: formatCreatedAt(namespace.Annotations),
	}
}

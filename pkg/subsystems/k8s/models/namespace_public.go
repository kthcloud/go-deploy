package models

import (
	"go-deploy/pkg/subsystems/k8s/keys"
	"go-deploy/utils/subsystemutils"
	v1 "k8s.io/api/core/v1"
)

type NamespacePublic struct {
	ID       string `bson:"id"`
	Name     string `bson:"name"`
	FullName string `bson:"fullName"`
}

func CreateNamespacePublicFromRead(namespace *v1.Namespace) *NamespacePublic {
	return &NamespacePublic{
		ID:       namespace.Labels[keys.ManifestLabelID],
		Name:     namespace.Labels[keys.ManifestLabelName],
		FullName: subsystemutils.GetPrefixedName(namespace.Labels[keys.ManifestLabelName]),
	}
}

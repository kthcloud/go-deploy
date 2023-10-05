package models

import (
	"go-deploy/pkg/subsystems/k8s/keys"
	v1 "k8s.io/api/core/v1"
	"time"
)

type SecretPublic struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Data      map[string][]byte `json:"data"`
	Type      string            `json:"type"`
	CreatedAt time.Time         `json:"createdAt"`
}

func (secret *SecretPublic) Created() bool {
	return secret.ID != ""
}

func CreateSecretPublicFromRead(secret *v1.Secret) *SecretPublic {
	return &SecretPublic{
		ID:        secret.Labels[keys.ManifestLabelID],
		Name:      secret.ObjectMeta.Name,
		Namespace: secret.ObjectMeta.Namespace,
		Type:      string(secret.Type),
		CreatedAt: formatCreatedAt(secret.Annotations),
	}
}

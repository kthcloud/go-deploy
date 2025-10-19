package models

import (
	"time"

	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/keys"
	v1 "k8s.io/api/core/v1"
)

type K8sResource interface {
	Created() bool
}

type EnvVar struct {
	Name  string `bson:"name"`
	Value string `bson:"value"`
}

type Volume struct {
	Name      string  `bson:"name"`
	PvcName   *string `bson:"pvcName"`
	MountPath string  `bson:"mountPath"`
	Init      bool    `bson:"init"`
}

type InitContainer struct {
	Name    string   `bson:"name"`
	Image   string   `bson:"image"`
	Command []string `bson:"command"`
	Args    []string `bson:"args"`
}

type Limits struct {
	CPU    string `bson:"cpu"`
	Memory string `bson:"memory"`
}

type Requests struct {
	CPU    string `bson:"cpu"`
	Memory string `bson:"memory"`
}

type Resources struct {
	Limits   Limits   `bson:"limits"`
	Requests Requests `bson:"requests"`
}

type ResourceClaim struct {
	Name                      string   `bson:"name"`
	Request                   []string `bson:"request,omitempty"`
	ResourceClaimName         *string  `bson:"resourceClaimName,omitempty"`
	ResourceClaimTemplateName *string  `bson:"resourceClaimTemplateName,omitempty"`
}

// ToK8sEnvVar converts an EnvVar to a v1.EnvVar.
func (envVar *EnvVar) ToK8sEnvVar() v1.EnvVar {
	return v1.EnvVar{
		Name:      envVar.Name,
		Value:     envVar.Value,
		ValueFrom: nil,
	}
}

// EnvVarFromK8s converts a v1.EnvVar to an EnvVar.
func EnvVarFromK8s(envVar *v1.EnvVar) EnvVar {
	return EnvVar{
		Name:  envVar.Name,
		Value: envVar.Value,
	}
}

// formatCreatedAt formats a Kubernetes manifest's creation timestamp to a time.Time.
func formatCreatedAt(annotations map[string]string) time.Time {
	created, ok := annotations[keys.AnnotationCreationTimestamp]
	if !ok {
		return time.Now()
	}

	createdAt, err := time.Parse("2006-01-02 15:04:05.000 -0700", created)
	if err != nil {
		return time.Now()
	}

	return createdAt
}

// clearSystemLabels removes system labels from a Kubernetes manifest.
func clearSystemLabels(labels map[string]string) map[string]string {
	delete(labels, keys.LabelDeployName)
	return labels
}

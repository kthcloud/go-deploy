package models

import (
	"go-deploy/pkg/subsystems/k8s/keys"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

type EnvVar struct {
	Name  string `bson:"name"`
	Value string `bson:"value"`
}

func (envVar *EnvVar) ToK8sEnvVar() v1.EnvVar {
	return v1.EnvVar{
		Name:      envVar.Name,
		Value:     envVar.Value,
		ValueFrom: nil,
	}
}

func EnvVarFromK8s(envVar *v1.EnvVar) EnvVar {
	return EnvVar{
		Name:  envVar.Name,
		Value: envVar.Value,
	}
}

type DeploymentPublic struct {
	ID          string   `bson:"id"`
	Name        string   `bson:"name"`
	Namespace   string   `bson:"namespace"`
	DockerImage string   `bson:"dockerImage"`
	EnvVars     []EnvVar `bson:"envVars"`
}

func (d *DeploymentPublic) Created() bool {
	return d.ID != ""
}

func CreateDeploymentPublicFromRead(deployment *appsv1.Deployment) *DeploymentPublic {
	var envs []EnvVar
	for _, env := range deployment.Spec.Template.Spec.Containers[0].Env {
		envs = append(envs, EnvVarFromK8s(&env))
	}

	return &DeploymentPublic{
		ID:          deployment.Labels[keys.ManifestLabelID],
		Name:        deployment.Labels[keys.ManifestLabelName],
		Namespace:   deployment.Namespace,
		DockerImage: deployment.Spec.Template.Spec.Containers[0].Image,
		EnvVars:     envs,
	}
}

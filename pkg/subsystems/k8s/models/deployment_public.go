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
	ID          string    `bson:"id"`
	Name        string    `bson:"name"`
	Namespace   string    `bson:"namespace"`
	DockerImage string    `bson:"dockerImage"`
	EnvVars     []EnvVar  `bson:"envVars"`
	Resources   Resources `bson:"resources"`
}

func (d *DeploymentPublic) Created() bool {
	return d.ID != ""
}

func CreateDeploymentPublicFromRead(deployment *appsv1.Deployment) *DeploymentPublic {
	var envs []EnvVar
	for _, env := range deployment.Spec.Template.Spec.Containers[0].Env {
		envs = append(envs, EnvVarFromK8s(&env))
	}

	limits := Limits{}
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		if deployment.Spec.Template.Spec.Containers[0].Resources.Limits != nil {
			if deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu() != nil {
				limits.CPU = deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().String()
			}
			if deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Memory() != nil {
				limits.Memory = deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().String()
			}
		}
	}

	requests := Requests{}
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		if deployment.Spec.Template.Spec.Containers[0].Resources.Requests != nil {
			if deployment.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu() != nil {
				requests.CPU = deployment.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().String()
			}
			if deployment.Spec.Template.Spec.Containers[0].Resources.Requests.Memory() != nil {
				requests.Memory = deployment.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().String()
			}
		}
	}

	return &DeploymentPublic{
		ID:          deployment.Labels[keys.ManifestLabelID],
		Name:        deployment.Labels[keys.ManifestLabelName],
		Namespace:   deployment.Namespace,
		DockerImage: deployment.Spec.Template.Spec.Containers[0].Image,
		EnvVars:     envs,
		Resources: Resources{
			Limits:   limits,
			Requests: requests,
		},
	}
}

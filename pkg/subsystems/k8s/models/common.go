package models

import v1 "k8s.io/api/core/v1"

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

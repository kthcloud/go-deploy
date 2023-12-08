package storage_manager

import "go-deploy/models/sys/deployment/subsystems"

type Subsystems struct {
	K8s subsystems.K8s `json:"k8s" bson:"k8s"`
}

type Volume struct {
	Name       string `bson:"name"`
	Init       bool   `bson:"init"`
	AppPath    string `bson:"appPath"`
	ServerPath string `bson:"serverPath"`
}

type Job struct {
	Name    string   `bson:"name"`
	Image   string   `bson:"image"`
	Command []string `bson:"command"`
	Args    []string `bson:"args"`
}

type InitContainer struct {
	Name    string   `bson:"name"`
	Image   string   `bson:"image"`
	Command []string `bson:"command"`
	Args    []string `bson:"args"`
}

type CreateParams struct {
	Zone string `json:"zone" bson:"zone"`
}

package models

import appsv1 "k8s.io/api/apps/v1"

type DeploymentStatus struct {
	Name                string
	Generation          int
	DesiredReplicas     int
	ReadyReplicas       int
	AvailableReplicas   int
	UnavailableReplicas int
}

func CreateDeploymentStatusFromRead(read *appsv1.Deployment) *DeploymentStatus {
	return &DeploymentStatus{
		Name:                read.Name,
		Generation:          int(read.Generation),
		DesiredReplicas:     int(read.Status.Replicas),
		ReadyReplicas:       int(read.Status.ReadyReplicas),
		AvailableReplicas:   int(read.Status.AvailableReplicas),
		UnavailableReplicas: int(read.Status.UnavailableReplicas),
	}
}

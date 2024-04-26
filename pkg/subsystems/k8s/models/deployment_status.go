package models

type DeploymentStatus struct {
	Generation          int
	DesiredReplicas     int
	ReadyReplicas       int
	AvailableReplicas   int
	UnavailableReplicas int
}

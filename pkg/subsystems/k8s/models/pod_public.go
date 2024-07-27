package models

import v1 "k8s.io/api/core/v1"

type PodPublic struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

func CreatePodPublicFromRead(pod v1.Pod) *PodPublic {
	return &PodPublic{
		Name:      pod.Name,
		Namespace: pod.Namespace,
	}
}

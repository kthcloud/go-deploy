package models

import corev1 "k8s.io/api/core/v1"

type NodePublic struct {
	Name string `json:"name"`
	CPU  struct {
		Total int `json:"total"`
	} `json:"cpu"`
	RAM struct {
		Total int `json:"total"`
	} `json:"ram"`
}

func CreateNodePublicFromGet(node *corev1.Node) NodePublic {
	return NodePublic{
		Name: node.Name,
		CPU: struct {
			Total int `json:"total"`
		}{
			Total: int(node.Status.Capacity.Cpu().MilliValue()) / 1000,
		},
		RAM: struct {
			Total int `json:"total"`
		}{
			Total: int(float64(node.Status.Capacity.Memory().Value() / 1024 / 1024 / 1024)),
		},
	}
}

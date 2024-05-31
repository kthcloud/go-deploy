package k8s

import (
	"context"
	"go-deploy/pkg/subsystems/k8s/models"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (client *Client) ListNodes() ([]models.NodePublic, error) {
	nodes, err := client.K8sClient.CoreV1().Nodes().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var res []models.NodePublic
	for _, node := range nodes.Items {
		res = append(res, models.CreateNodePublicFromGet(&node))
	}

	return res, nil
}

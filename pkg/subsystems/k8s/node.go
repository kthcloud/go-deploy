package k8s

import (
	"context"
	"fmt"
	"go-deploy/pkg/subsystems/k8s/models"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReadNode returns the node with the given name
func (client *Client) ReadNode(name string) (*models.NodePublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read k8s node %s. details: %w", name, err)
	}

	if name == "" {
		return nil, fmt.Errorf("no name supplied when reading k8s node")
	}

	node, err := client.K8sClient.CoreV1().Nodes().Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil

		}

		return nil, makeError(err)
	}

	return models.CreateNodePublicFromGet(node), nil
}

func (client *Client) ListNodes() ([]models.NodePublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to list k8s nodes. details: %w", err)
	}

	nodes, err := client.K8sClient.CoreV1().Nodes().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, makeError(err)
	}

	var res []models.NodePublic
	for _, node := range nodes.Items {
		res = append(res, *models.CreateNodePublicFromGet(&node))
	}

	return res, nil
}

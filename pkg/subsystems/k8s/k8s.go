package k8s

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
)

type Client struct {
	K8sClient *kubernetes.Clientset
	Namespace string
}

func New(k8sClient *kubernetes.Clientset, namespace string) (*Client, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to create k8s client. details: %w", err)
	}

	client := Client{
		K8sClient: k8sClient,
		Namespace: namespace,
	}

	return &client, nil
}

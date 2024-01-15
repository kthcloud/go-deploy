package k8s

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
)

// Client is a wrapper around the Kubernetes client.
type Client struct {
	K8sClient *kubernetes.Clientset
	Namespace string
}

// New creates a new Kubernetes wrapper client.
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

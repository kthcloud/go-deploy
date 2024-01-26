package k8s

import (
	"fmt"
	"go-deploy/pkg/imp/kubevirt/kubevirt"
	"k8s.io/client-go/kubernetes"
)

// ClientConf is the configuration for the Kubernetes wrapper client.
type ClientConf struct {
	K8sClient     *kubernetes.Clientset
	VirtK8sClient *kubevirt.Clientset
	Namespace     string
}

// Client is a wrapper around the Kubernetes client.
type Client struct {
	K8sClient     *kubernetes.Clientset
	VirtK8sClient *kubevirt.Clientset
	Namespace     string
}

// New creates a new Kubernetes wrapper client.
func New(conf *ClientConf) (*Client, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to create k8s client. details: %w", err)
	}

	client := Client{
		K8sClient:     conf.K8sClient,
		VirtK8sClient: conf.VirtK8sClient,
		Namespace:     conf.Namespace,
	}

	return &client, nil
}

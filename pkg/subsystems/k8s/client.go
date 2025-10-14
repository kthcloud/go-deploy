package k8s

import (
	"fmt"
	"github.com/kthcloud/go-deploy/pkg/imp/kubevirt/kubevirt"
	"k8s.io/client-go/kubernetes"
)

// ClientConf is the configuration for the Kubernetes wrapper client.
type ClientConf struct {
	K8sClient         *kubernetes.Clientset
	KubeVirtK8sClient *kubevirt.Clientset
	Namespace         string
}

// Client is a wrapper around the Kubernetes client.
type Client struct {
	K8sClient         *kubernetes.Clientset
	KubeVirtK8sClient *kubevirt.Clientset
	Namespace         string
}

// New creates a new Kubernetes wrapper client.
func New(conf *ClientConf) (*Client, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to create k8s client. details: %w", err)
	}

	client := Client{
		K8sClient:         conf.K8sClient,
		KubeVirtK8sClient: conf.KubeVirtK8sClient,
		Namespace:         conf.Namespace,
	}

	return &client, nil
}

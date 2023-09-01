package k8s

import (
	"encoding/base64"
	"k8s.io/client-go/kubernetes"
)

type Client struct {
	K8sClient *kubernetes.Clientset
}

type ClientConf struct {
	K8sAuth string
}

func New(k8sClient *kubernetes.Clientset) (*Client, error) {
	//makeError := func(err error) error {
	//	return fmt.Errorf("failed to create k8s client. details: %w", err)
	//}

	client := Client{
		K8sClient: k8sClient,
	}

	return &client, nil
}

func createConfigFromB64(b64Config string) []byte {
	configB64 := b64Config
	config, _ := base64.StdEncoding.DecodeString(configB64)
	return config
}

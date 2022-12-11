package k8s

import (
	"encoding/base64"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	K8sClient *kubernetes.Clientset
}

type ClientConf struct {
	K8sAuth string
}

func New(config *ClientConf) (*Client, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create npm client. details: %s", err)
	}

	configData := createConfigFromB64(config.K8sAuth)
	kubeConfig, err := clientcmd.RESTConfigFromKubeConfig(configData)
	if err != nil {
		return nil, makeError(err)
	}

	k8sClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, makeError(err)
	}

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

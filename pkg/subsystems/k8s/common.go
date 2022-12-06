package k8s

import (
	"encoding/base64"
	"fmt"
	"go-deploy/pkg/conf"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
)

func createConfigFromB64() []byte {
	configB64 := conf.Env.K8s.Config
	config, _ := base64.StdEncoding.DecodeString(configB64)
	return config
}

var client *kubernetes.Clientset

func Setup() {
	makeError := func(err error) error {
		return fmt.Errorf("failed to setup k8s. details: %s", err)
	}

	configData := createConfigFromB64()
	kubeConfig, err := clientcmd.RESTConfigFromKubeConfig(configData)
	if err != nil {
		log.Fatalln(makeError(err))
	}

	newClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		log.Fatalln(makeError(err))
	}
	client = newClient
}

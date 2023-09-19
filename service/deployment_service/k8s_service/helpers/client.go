package helpers

import (
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/models/sys/deployment/subsystems"
	"go-deploy/models/sys/enviroment"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/k8s"
)

type Client struct {
	UpdateDB func(string, string, interface{}) error
	// subsystem client
	SsClient  *k8s.Client
	Zone      *enviroment.DeploymentZone
	K8s       *subsystems.K8s
	Namespace string
}

func New(k8s *subsystems.K8s, zoneName, namespace string) (*Client, error) {
	makeError := func(err error) error {
		return fmt.Errorf("error creating k8s client in deployment helper client. details: %w", err)
	}

	zone := conf.Env.Deployment.GetZone(zoneName)
	if zone == nil {
		return nil, makeError(fmt.Errorf("zone %s not found", zoneName))
	}

	k8sClient, err := withClient(zone, namespace)
	if err != nil {
		return nil, makeError(err)
	}

	return &Client{
		UpdateDB: func(id, key string, data interface{}) error {
			return deploymentModel.New().UpdateSubsystemByID(id, "k8s", key, data)
		},
		SsClient:  k8sClient,
		Zone:      zone,
		K8s:       k8s,
		Namespace: namespace,
	}, nil
}

func withClient(zone *enviroment.DeploymentZone, namespace string) (*k8s.Client, error) {
	client, err := k8s.New(zone.Client, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client. details: %w", err)
	}

	if namespace != "" {
		namespaceCreated, err := client.NamespaceCreated(namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to check if namespace %s is created. details: %w", namespace, err)
		}

		if !namespaceCreated {
			return nil, fmt.Errorf("no such namespace %s", namespace)
		}
	}

	return client, nil
}

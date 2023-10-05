package helpers

import (
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/models/sys/deployment/subsystems"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/harbor"
)

type Client struct {
	UpdateDB func(string, string, interface{}) error
	// subsystem client
	SsClient *harbor.Client
	// subsystem
	SS *subsystems.Harbor
}

func New(harbor *subsystems.Harbor) (*Client, error) {
	makeError := func(err error) error {
		return fmt.Errorf("error creating harbor client in deployment helper client. details: %w", err)
	}

	harborClient, err := withClient()
	if err != nil {
		return nil, makeError(err)
	}

	return &Client{
		UpdateDB: func(id, key string, data interface{}) error {
			return deploymentModel.New().UpdateSubsystemByID(id, "harbor", key, data)
		},
		SsClient: harborClient,
		SS:       harbor,
	}, nil
}

func withClient() (*harbor.Client, error) {
	return harbor.New(&harbor.ClientConf{
		URL:      conf.Env.Harbor.URL,
		Username: conf.Env.Harbor.User,
		Password: conf.Env.Harbor.Password,
	})
}

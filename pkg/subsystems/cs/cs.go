package cs

import (
	"fmt"

	"go-deploy/pkg/imp/cloudstack"
)

type Client struct {
	CsClient  *cloudstack.CloudStackClient
	NetworkID string
	ProjectID string
	ZoneID    string
}

type ClientConf struct {
	URL    string
	ApiKey string
	Secret string

	NetworkID string
	ProjectID string
	ZoneID    string
}

func New(config *ClientConf) (*Client, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to create cloudstack client. details: %s", err)
	}

	csClient := cloudstack.NewAsyncClient(
		config.URL,
		config.ApiKey,
		config.Secret,
		true,
	)

	client := Client{
		CsClient:  csClient,
		NetworkID: config.NetworkID,
		ProjectID: config.ProjectID,
		ZoneID:    config.ZoneID,
	}

	return &client, nil
}

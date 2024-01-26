package rancher

import (
	"fmt"
	"github.com/rancher/go-rancher/client"
)

type ClientConf struct {
	URL    string
	ApiKey string
	Secret string
}

type Client struct {
	RancherClient *client.RancherClient

	URL    string
	ApiKey string
	Secret string
}

func New(config *ClientConf) (*Client, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create rancher client. details: %w", err)
	}

	rancherClient, err := client.NewRancherClient(&client.ClientOpts{
		Url:       config.URL,
		AccessKey: config.ApiKey,
		SecretKey: config.Secret,
	})
	if err != nil {
		return nil, makeError(err)
	}

	return &Client{
		RancherClient: rancherClient,
		URL:           config.URL,
		ApiKey:        config.ApiKey,
		Secret:        config.Secret,
	}, nil
}

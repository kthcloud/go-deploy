package cs

import (
	"fmt"

	"github.com/apache/cloudstack-go/v2/cloudstack"
)

type Client struct {
	CSClient *cloudstack.CloudStackClient
}

type ClientConf struct {
	ApiUrl    string
	ApiKey    string
	SecretKey string
}

func New(config *ClientConf) (*Client, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to create cloudstack client. details: %s", err)
	}

	csClient := cloudstack.NewAsyncClient(
		config.ApiUrl,
		config.ApiKey,
		config.SecretKey,
		true,
	)

	client := Client{
		CSClient: csClient,
	}

	return &client, nil
}

package cs

import (
	"github.com/apache/cloudstack-go/v2/cloudstack"
)

type Client struct {
	CSClient *cloudstack.CloudStackClient
	ZoneID   string
}

type ClientConf struct {
	ApiUrl    string
	ApiKey    string
	ApiSecret string
	ZoneID    string
}

func New(config *ClientConf) (*Client, error) {
	csClient := cloudstack.NewAsyncClient(
		config.ApiUrl,
		config.ApiKey,
		config.ApiSecret,
		true,
	)

	client := Client{
		CSClient: csClient,
		ZoneID:   config.ZoneID,
	}

	return &client, nil
}

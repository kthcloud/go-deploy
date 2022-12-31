package pdns

import (
	"fmt"
	"github.com/joeig/go-powerdns/v3"
)

type Client struct {
	apiUrl     string
	apiKey     string
	PdnsClient *powerdns.Client
	Zone       string
}

type ClientConf struct {
	ApiUrl string
	ApiKey string
	Zone   string
}

func New(config *ClientConf) (*Client, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to create pdns client. details: %s", err)
	}

	pdnsClient := powerdns.NewClient(config.ApiUrl, "localhost", map[string]string{"X-API-Key": config.ApiKey}, nil)

	client := Client{
		apiUrl:     config.ApiUrl,
		apiKey:     config.ApiKey,
		PdnsClient: pdnsClient,
		Zone:       config.Zone,
	}

	return &client, nil
}

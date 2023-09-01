package harbor

import (
	"fmt"
	"github.com/mittwald/goharbor-client/v5/apiv2"
	"github.com/mittwald/goharbor-client/v5/apiv2/pkg/config"
)

type Client struct {
	apiUrl       string
	username     string
	password     string
	HarborClient *apiv2.RESTClient
}

type ClientConf struct {
	ApiUrl   string
	Username string
	Password string
}

func New(config *ClientConf) (*Client, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create harbor client. details: %w", err)
	}

	harborClient, err := createHarborClient(config.ApiUrl, config.Username, config.Password)
	if err != nil {
		return nil, makeError(err)
	}

	client := Client{
		apiUrl:       config.ApiUrl,
		username:     config.Username,
		password:     config.Password,
		HarborClient: harborClient,
	}

	return &client, nil
}

func createHarborClient(apiUrl, username, password string) (*apiv2.RESTClient, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create harbor client. details: %w", err.Error())
	}

	client, err := apiv2.NewRESTClientForHost(apiUrl, username, password, &config.Options{})
	if err != nil {
		return nil, makeError(err)
	}

	return client, nil
}

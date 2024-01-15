package harbor

import (
	"fmt"
	"github.com/mittwald/goharbor-client/v5/apiv2"
	"github.com/mittwald/goharbor-client/v5/apiv2/pkg/config"
)

// Client is a wrapper around the Harbor API client.
type Client struct {
	url      string
	username string
	password string

	Project      string
	HarborClient *apiv2.RESTClient
}

// ClientConf is a configuration struct for the Harbor client.
type ClientConf struct {
	URL      string
	Username string
	Password string
	Project  string
}

// New creates a new Harbor client.
func New(config *ClientConf) (*Client, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create harbor client. details: %w", err)
	}

	harborClient, err := createHarborClient(config.URL, config.Username, config.Password)
	if err != nil {
		return nil, makeError(err)
	}

	client := Client{
		url:          config.URL,
		username:     config.Username,
		password:     config.Password,
		Project:      config.Project,
		HarborClient: harborClient,
	}

	return &client, nil
}

// createHarborClient is a helper function to create a Harbor client.
func createHarborClient(apiUrl, username, password string) (*apiv2.RESTClient, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create harbor client. details: %w", err)
	}

	client, err := apiv2.NewRESTClientForHost(apiUrl, username, password, &config.Options{})
	if err != nil {
		return nil, makeError(err)
	}

	return client, nil
}

package harbor

import (
	"fmt"
	"github.com/kthcloud/go-deploy/pkg/imp/harbor/harbor"
)

// Client is a wrapper around the Harbor API client.
type Client struct {
	url      string
	username string
	password string

	Project      string
	HarborClient *harbor.ClientSet
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
func createHarborClient(apiUrl, username, password string) (*harbor.ClientSet, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create harbor client. details: %w", err)
	}

	client, err := harbor.NewClientSet(&harbor.ClientSetConfig{
		URL:      apiUrl,
		Insecure: true,
		Username: username,
		Password: password,
	})

	if err != nil {
		return nil, makeError(err)
	}

	return client, nil
}

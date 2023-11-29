package cs

import (
	"fmt"
	"go-deploy/pkg/subsystems/cs/models"
)

func (client *Client) ReadHostByName(name string) (*models.HostPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read host %s. details: %w", name, err)
	}

	host, _, err := client.CsClient.Host.GetHostByName(name)
	if err != nil {
		return nil, makeError(err)
	}

	if host == nil {
		return nil, nil
	}

	return models.CreateHostPublicFromGet(host), nil
}

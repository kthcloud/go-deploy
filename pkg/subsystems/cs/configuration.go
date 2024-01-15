package cs

import (
	"fmt"
	"go-deploy/pkg/subsystems/cs/models"
)

// ReadConfiguration reads the configuration from CloudStack.
func (client *Client) ReadConfiguration() (*models.ConfigurationPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read configuration. details: %w", err)
	}

	params := client.CsClient.Configuration.NewListConfigurationsParams()
	params.SetName("cpu.overprovisioning.factor")
	params.SetCategory("Advanced")

	configurationList, err := client.CsClient.Configuration.ListConfigurations(params)
	if err != nil {
		return nil, makeError(err)
	}

	if configurationList == nil {
		return nil, nil
	}

	return models.CreateConfigurationPublicFromGet(configurationList), nil
}

package models

import (
	"go-deploy/pkg/imp/cloudstack"
	"strconv"
)

type ConfigurationPublic struct {
	OverProvisioningFactor int
}

func CreateConfigurationPublicFromGet(configurations *cloudstack.ListConfigurationsResponse) *ConfigurationPublic {
	res := &ConfigurationPublic{}

	for _, configuration := range configurations.Configurations {
		if configuration.Name == "cpu.overprovisioning.factor" {
			val, err := strconv.Atoi(configuration.Value)
			if err != nil {
				val = 1
			}

			res.OverProvisioningFactor = val
		}
	}

	return res
}

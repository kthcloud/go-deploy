package intializer

import (
	"errors"

	"github.com/google/uuid"
	body "github.com/kthcloud/go-deploy/api/openapi/v2"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/service"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
)

// EnsureSystemDeploymentsExists ensures that the deployments related to the system are created.
// This includes the fallback deployment, which is used by other deployments.
func EnsureSystemDeploymentsExists() error {
	// Fallback-disabled deployment
	for _, zone := range config.Config.Zones {
		if config.Config.Deployment.Fallback.Disabled.Name == "" {
			return errors.New("fallback deployment name not set")
		}

		err := service.V2().Deployments().Create(uuid.NewString(), "system", &body.DeploymentCreate{
			Name:     config.Config.Deployment.Fallback.Disabled.Name,
			CpuCores: new(float64(1)),
			RAM:      new(float64(1)),
			Replicas: new(5),
			Envs: []body.Env{
				{
					Name:  "TYPE",
					Value: "disabled",
				},
			},
			Image:           new(config.Config.Registry.PlaceholderImage),
			HealthCheckPath: new(""),
			Zone:            new(zone.Name),
		})
		if err != nil {
			if !errors.Is(err, sErrors.ErrNonUniqueField) {
				return err
			}
		}

		// Deployment either already exists or was created
		// Ensure the owner is "system"
		deployment, err := service.V2().Deployments().GetByName(config.Config.Deployment.Fallback.Disabled.Name)
		if err != nil {
			return err
		}

		if deployment == nil {
			return errors.New("deployment not found")
		}

		if deployment.OwnerID != "system" {
			return errors.New("deployment owner is not system")
		}
	}

	return nil
}

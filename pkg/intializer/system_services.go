package intializer

import (
	"errors"
	"github.com/google/uuid"
	"github.com/kthcloud/go-deploy/dto/v2/body"
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
			CpuCores: floatPtr(1),
			RAM:      floatPtr(1),
			Replicas: intPtr(5),
			Envs: []body.Env{
				{
					Name:  "TYPE",
					Value: "disabled",
				},
			},
			Image:           strPtr(config.Config.Registry.PlaceholderImage),
			HealthCheckPath: strPtr(""),
			Zone:            strPtr(zone.Name),
		})
		if err != nil {
			if !errors.Is(err, sErrors.NonUniqueFieldErr) {
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

func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}

func strPtr(s string) *string {
	return &s
}

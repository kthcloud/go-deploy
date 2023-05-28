package deployment_service

import (
	"context"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/utils/subsystemutils"
	"log"
)

func GetLogs(userID, deploymentID string, handler func(string), isAdmin bool) (context.Context, error) {
	deployment, err := GetByID(userID, deploymentID, isAdmin)
	if err != nil {
		return nil, err
	}

	if deployment == nil {
		log.Println("deployment", deploymentID, "not found when getting logs. assuming it was deleted")
		return nil, nil
	}

	if !deployment.Ready() {
		log.Println("deployment", deploymentID, "not ready when getting logs")
		return nil, nil
	}

	ctx := context.Background()

	client, err := k8s.New(conf.Env.K8s.Client)
	if err != nil {
		return nil, err
	}

	subsystem := deployment.Subsystems.K8s
	err = client.GetLogStream(ctx, subsystemutils.GetPrefixedName(subsystem.Namespace.Name), subsystem.Deployment.Name, handler)
	if err != nil {
		ctx.Done()
		return nil, err
	}

	return ctx, nil
}

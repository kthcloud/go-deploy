package deployment_service

import (
	"context"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/utils/subsystemutils"
)

func GetLogs(userID, deploymentID string, handler func(string)) (context.Context, error) {
	deployment, err := GetByFullID(userID, deploymentID)
	if err != nil {
		return nil, err
	}

	if deployment == nil {
		return nil, nil
	}

	if !deployment.Ready() {
		return nil, nil
	}

	ctx := context.Background()

	client, err := k8s.New(conf.Env.K8s.Client)
	if err != nil {
		return nil, err
	}

	err = client.GetLogStream(ctx, subsystemutils.GetPrefixedName(deployment.Name), handler)
	if err != nil {
		ctx.Done()
		return nil, err
	}

	return ctx, nil
}

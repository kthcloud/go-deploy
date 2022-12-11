package deployment_service

import (
	"context"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/utils/subsystemutils"
)

func GetLogs(userID, deploymentID string, handler func(string)) (context.Context, error) {
	deployment, err := Get(userID, deploymentID)
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

	client, err := k8s.New(&k8s.ClientConf{K8sAuth: conf.Env.K8s.Config})
	if err != nil {
		return nil, err
	}

	err = client.GetLogStream(ctx, subsystemutils.GetPrefixedName(deployment.Name), deployment.Name, handler)
	if err != nil {
		ctx.Done()
		return nil, err
	}

	return ctx, nil
}

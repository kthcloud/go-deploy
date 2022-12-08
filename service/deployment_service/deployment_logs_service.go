package deployment_service

import (
	"context"
	"go-deploy/pkg/subsystems/k8s"
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

	err = k8s.GetLogStream(ctx, deployment.Name, handler)
	if err != nil {
		ctx.Done()
		return nil, err
	}

	return ctx, nil
}

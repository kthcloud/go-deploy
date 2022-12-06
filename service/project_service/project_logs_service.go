package project_service

import (
	"context"
	"go-deploy/pkg/subsystems/k8s"
)

func GetLogs(userId, projectId string, handler func(string)) (context.Context, error) {
	project, err := Get(userId, projectId)
	if err != nil {
		return nil, err
	}

	if project == nil {
		return nil, nil
	}

	ctx := context.Background()

	err = k8s.GetLogStream(ctx, project.Name, handler)
	if err != nil {
		ctx.Done()
		return nil, err
	}

	return ctx, nil
}

package deployment_service

import (
	"go-deploy/pkg/status_codes"
)

func GetStatusByID(deploymentID string) (int, string, error) {
	deployment, err := GetByID(deploymentID)
	if err != nil {
		return -1, "Unknown", err
	}

	if deployment == nil {
		return status_codes.ResourceNotFound, status_codes.GetMsg(status_codes.ResourceNotFound), nil
	}

	if deployment.BeingDeleted {
		return status_codes.ResourceNotReady, status_codes.GetMsg(status_codes.ResourceNotReady), nil
	}

	if deployment.BeingCreated {
		return status_codes.ResourceNotReady, status_codes.GetMsg(status_codes.ResourceNotReady), nil
	}

	// TODO: fetch status from k8s

	return status_codes.ResourceCreated, status_codes.GetMsg(status_codes.ResourceRunning), nil
}

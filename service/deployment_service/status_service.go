package deployment_service

import (
	"go-deploy/pkg/status_codes"
)

func GetStatusByID(userID, deploymentID string) (int, string, error) {
	deployment, err := GetByID(userID, deploymentID)
	if err != nil {
		return -1, "Unknown", err
	}

	if deployment == nil {
		return status_codes.ResourceNotFound, status_codes.GetMsg(status_codes.ResourceNotFound), nil
	}

	if deployment.BeingDeleted {
		return status_codes.ResourceBeingDeleted, status_codes.GetMsg(status_codes.ResourceBeingDeleted), nil
	}

	if deployment.BeingCreated {
		return status_codes.ResourceBeingCreated, status_codes.GetMsg(status_codes.ResourceBeingCreated), nil
	}

	// TODO: fetch status from k8s

	return status_codes.ResourceCreated, status_codes.GetMsg(status_codes.ResourceRunning), nil
}

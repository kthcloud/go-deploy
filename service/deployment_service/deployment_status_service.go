package deployment_service

import (
	"go-deploy/models/dto"
	"go-deploy/pkg/status_codes"
)

func GetStatusByID(userID, deploymentID string) (int, *dto.DeploymentStatus, error) {
	deployment, err := Get(userID, deploymentID)
	if err != nil {
		return -1, nil, err
	}

	if deployment == nil {
		return status_codes.DeploymentNotFound, &dto.DeploymentStatus{
			Status: status_codes.GetMsg(status_codes.DeploymentNotFound),
		}, nil
	}

	if deployment.BeingDeleted {
		return status_codes.DeploymentBeingDeleted, &dto.DeploymentStatus{
			Status: status_codes.GetMsg(status_codes.DeploymentBeingDeleted),
		}, nil
	}

	if deployment.BeingCreated {
		return status_codes.DeploymentBeingCreated, &dto.DeploymentStatus{
			Status: status_codes.GetMsg(status_codes.DeploymentBeingCreated),
		}, nil
	}

	return status_codes.DeploymentCreated, &dto.DeploymentStatus{
		Status: status_codes.GetMsg(status_codes.DeploymentCreated),
	}, nil
}

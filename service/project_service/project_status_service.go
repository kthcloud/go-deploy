package project_service

import (
	"deploy-api-go/models/dto"
	"deploy-api-go/pkg/status_codes"
)

func GetStatusByID(userId, projectId string) (int, *dto.ProjectStatus, error) {
	project, err := Get(userId, projectId)
	if err != nil {
		return -1, nil, err
	}

	if project == nil {
		return status_codes.ProjectNotFound, &dto.ProjectStatus{
			Status: status_codes.GetMsg(status_codes.ProjectNotFound),
		}, nil
	}

	if project.BeingDeleted {
		return status_codes.ProjectBeingDeleted, &dto.ProjectStatus{
			Status: status_codes.GetMsg(status_codes.ProjectBeingDeleted),
		}, nil
	}

	if project.BeingCreated {
		return status_codes.ProjectBeingCreated, &dto.ProjectStatus{
			Status: status_codes.GetMsg(status_codes.ProjectBeingCreated),
		}, nil
	}

	return status_codes.ProjectCreated, &dto.ProjectStatus{
		Status: status_codes.GetMsg(status_codes.ProjectCreated),
	}, nil
}

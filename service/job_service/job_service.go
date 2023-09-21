package job_service

import (
	"fmt"
	"go-deploy/models/dto/query"
	jobModel "go-deploy/models/sys/job"
	"go-deploy/service"
)

func Create(id, userID, jobType string, args map[string]interface{}) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create job. details: %w", err)
	}

	err := jobModel.New().Create(id, userID, jobType, args)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func GetByID(jobID string, auth *service.AuthInfo) (*jobModel.Job, error) {
	client := jobModel.New()
	if !auth.IsAdmin {
		client.AddRestrictedUser(&auth.UserID)
	}

	return client.GetByID(jobID)
}

func GetMany(allUsers bool, userID, jobType *string, status *string, auth *service.AuthInfo, pagination *query.Pagination) ([]jobModel.Job, error) {
	client := jobModel.New()

	if pagination != nil {
		client.AddPagination(pagination.Page, pagination.PageSize)
	}

	if userID != nil {
		if *userID != auth.UserID && !auth.IsAdmin {
			return nil, nil
		}
		client.AddRestrictedUser(userID)
	} else if !allUsers || (allUsers && !auth.IsAdmin) {
		client.AddRestrictedUser(&auth.UserID)
	}

	return client.GetMany(jobType, status)
}

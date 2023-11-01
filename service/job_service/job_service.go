package job_service

import (
	"fmt"
	"go-deploy/models/dto/body"
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

func ExistsAuth(id string, auth *service.AuthInfo) (bool, error) {
	client := jobModel.New()

	if !auth.IsAdmin {
		client.RestrictToUser(auth.UserID)
	}

	return client.ExistsByID(id)
}

func GetByIdAuth(jobID string, auth *service.AuthInfo) (*jobModel.Job, error) {
	client := jobModel.New()
	if !auth.IsAdmin {
		client.RestrictToUser(auth.UserID)
	}

	return client.GetByID(jobID)
}

func ListAuth(allUsers bool, userID, jobType *string, status *string, auth *service.AuthInfo, pagination *query.Pagination) ([]jobModel.Job, error) {
	client := jobModel.New()

	if pagination != nil {
		client.AddPagination(pagination.Page, pagination.PageSize)
	}

	if userID != nil {
		if *userID != auth.UserID && !auth.IsAdmin {
			return nil, nil
		}
		client.RestrictToUser(*userID)
	} else if !allUsers || (allUsers && !auth.IsAdmin) {
		client.RestrictToUser(auth.UserID)
	}

	return client.GetMany(jobType, status)
}

func UpdateAuth(id string, jobUpdateDTO *body.JobUpdate, auth *service.AuthInfo) (*jobModel.Job, error) {
	client := jobModel.New()

	if !auth.IsAdmin {
		client.RestrictToUser(auth.UserID)
	}

	var params jobModel.UpdateParams
	params.FromDTO(jobUpdateDTO)

	err := client.UpdateWithParams(id, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to update job. details: %w", err)
	}

	return client.GetByID(id)
}

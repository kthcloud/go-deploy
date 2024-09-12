package v2

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/dto/v2/query"
	"github.com/kthcloud/go-deploy/dto/v2/uri"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/app/status_codes"
	"github.com/kthcloud/go-deploy/pkg/sys"
	"github.com/kthcloud/go-deploy/service"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
	"github.com/kthcloud/go-deploy/service/v2/jobs/opts"
	v12 "github.com/kthcloud/go-deploy/service/v2/utils"
)

// GetJob
// @Summary GetJob job by id
// @Description GetJob job by id
// @Tags Job
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param jobId path string true "Job ID"
// @Success 200 {object} body.JobRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/jobs/{jobId} [get]
func GetJob(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.JobGet
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	job, err := service.V2(auth).Jobs().Get(requestURI.JobID)
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if job == nil {
		context.NotFound("Job not found")
		return
	}

	context.JSONResponse(200, job.ToDTO(jobStatusMessage(job.Status)))
}

// ListJobs
// @Summary List jobs
// @Description List jobs
// @Tags Job
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param all query bool false "List all"
// @Param userId query string false "Filter by user ID"
// @Param type query string false "Filter by type"
// @Param status query string false "Filter by status"
// @Param page query int false "Page number"
// @Param pageSize query int false "Number of items per page"
// @Success 200 {array} body.JobRead
// @Router /v2/jobs [get]
func ListJobs(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.JobList
	if err := context.GinContext.ShouldBindQuery(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	jobList, err := service.V2(auth).Jobs().List(opts.ListOpts{
		Pagination:      v12.GetOrDefaultPagination(requestQuery.Pagination),
		SortBy:          v12.GetOrDefaultSortBy(requestQuery.SortBy),
		UserID:          requestQuery.UserID,
		All:             requestQuery.All,
		JobTypes:        requestQuery.Types,
		ExcludeJobTypes: requestQuery.ExcludeTypes,
		Status:          requestQuery.Status,
		ExcludeStatus:   requestQuery.ExcludeStatus,
	})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if jobList == nil {
		context.Ok(make([]interface{}, 0))
		return
	}

	var jobsDTO []body.JobRead
	for _, job := range jobList {
		jobsDTO = append(jobsDTO, job.ToDTO(jobStatusMessage(job.Status)))
	}

	context.Ok(jobsDTO)
}

// UpdateJob
// @Summary Update job
// @Description Update job. Only allowed for admins.
// @Tags Job
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param jobId path string true "Job ID"
// @Param body body body.JobUpdate true "Job update"
// @Success 200 {object} body.JobRead
// @Router /v2/jobs/{jobId} [post]
func UpdateJob(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.JobUpdate
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	var request body.JobUpdate
	if err := context.GinContext.ShouldBindJSON(&request); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	updated, err := service.V2(auth).Jobs().Update(requestURI.JobID, &request)
	if err != nil {
		if errors.Is(err, sErrors.ForbiddenErr) {
			context.Forbidden("User is not allowed to update jobs")
			return
		}

		if errors.Is(err, sErrors.JobNotFoundErr) {
			context.NotFound("Job not found")
			return
		}

		context.ServerError(err, InternalError)
		return
	}

	if updated == nil {
		context.NotFound("Job not found")
		return
	}

	context.Ok(updated.ToDTO(jobStatusMessage(updated.Status)))
}

// jobStatusMessage is a helper function to get a parsed status message for a job
func jobStatusMessage(status string) string {
	switch status {
	case model.JobStatusPending:
		return status_codes.GetMsg(status_codes.JobPending)
	case model.JobStatusRunning:
		return status_codes.GetMsg(status_codes.JobRunning)
	case model.JobStatusCompleted:
		return status_codes.GetMsg(status_codes.JobFinished)
	case model.JobStatusFailed:
		return status_codes.GetMsg(status_codes.JobFailed)
	case model.JobStatusTerminated:
		return status_codes.GetMsg(status_codes.JobTerminated)

	default:
		return status_codes.GetMsg(status_codes.Unknown)
	}
}

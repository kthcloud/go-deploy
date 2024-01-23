package v1_job

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/v1/body"
	"go-deploy/models/dto/v1/query"
	"go-deploy/models/dto/v1/uri"
	jobModels "go-deploy/models/sys/job"
	"go-deploy/pkg/app/status_codes"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
	v12 "go-deploy/service/v1/common"
	"go-deploy/service/v1/jobs/opts"
)

// List
// @Summary Get list of jobs
// @Description Get list of jobs
// @Tags Job
// @Accept  json
// @Produce  json
// @Param all query bool false "Get all"
// @Param userId query string false "Filter by user id"
// @Param type query string false "Filter by type"
// @Param status query string false "Filter by status"
// @Param page query int false "Page number"
// @Param pageSize query int false "Number of items per page"
// @Success 200 {array} body.JobRead
// @Router /job [get]
func List(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.JobList
	if err := context.GinContext.ShouldBindQuery(&requestQuery); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	jobList, err := service.V1(auth).Jobs().List(opts.ListOpts{
		Pagination: v12.GetOrDefaultPagination(requestQuery.Pagination),
		SortBy:     v12.GetOrDefaultSortBy(requestQuery.SortBy),
		All:        requestQuery.All,
		UserID:     requestQuery.UserID,
		JobType:    requestQuery.Type,
		Status:     requestQuery.Status,
	})
	if err != nil {
		context.ServerError(err, v1.InternalError)
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

// Get
// @Summary Get job by id
// @Description Get job by id
// @Tags Job
// @Accept  json
// @Produce  json
// @Param jobId path string true "Job ID"
// @Success 200 {object} body.JobRead
// @Router /job/{id} [get]
func Get(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.JobGet
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	job, err := service.V1(auth).Jobs().Get(requestURI.JobID)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if job == nil {
		context.NotFound("Job not found")
		return
	}

	context.JSONResponse(200, job.ToDTO(jobStatusMessage(job.Status)))
}

// Update
// @Summary Update job
// @Description Update job
// @Tags Job
// @Accept  json
// @Produce  json
// @Param jobId path string true "Job ID"
// @Param body body body.JobUpdate true "Job update"
// @Success 200 {object} body.JobRead
// @Router /job/{id} [post]
func Update(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.JobUpdate
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	var request body.JobUpdate
	if err := context.GinContext.ShouldBindJSON(&request); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	updated, err := service.V1(auth).Jobs().Update(requestURI.JobID, &request)
	if err != nil {
		if errors.Is(err, sErrors.ForbiddenErr) {
			context.Forbidden("User is not allowed to update jobs")
			return
		}

		if errors.Is(err, sErrors.JobNotFoundErr) {
			context.NotFound("Job not found")
			return
		}

		context.ServerError(err, v1.InternalError)
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
	case jobModels.StatusPending:
		return status_codes.GetMsg(status_codes.JobPending)
	case jobModels.StatusRunning:
		return status_codes.GetMsg(status_codes.JobRunning)
	case jobModels.StatusCompleted:
		return status_codes.GetMsg(status_codes.JobFinished)
	case jobModels.StatusFailed:
		return status_codes.GetMsg(status_codes.JobFailed)
	case jobModels.StatusTerminated:
		return status_codes.GetMsg(status_codes.JobTerminated)

	// deprecated
	case jobModels.StatusFinished:
		return status_codes.GetMsg(status_codes.JobFinished)
	default:
		return status_codes.GetMsg(status_codes.Unknown)
	}
}

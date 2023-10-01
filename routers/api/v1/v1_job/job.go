package v1_job

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/models/dto/uri"
	jobModel "go-deploy/models/sys/job"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/job_service"
	"net/http"
)

// Get
// @Summary Get job by id
// @Description Get job by id
// @Tags Job
// @Accept  json
// @Produce  json
// @Param jobId path string true "Job ID"
// @Success 200 {object} body.JobRead
// @Router /api/v1/job/{id} [get]
func Get(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.JobGet
	if err := context.GinContext.BindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	job, err := job_service.GetByID(requestURI.JobID, auth)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if job == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, fmt.Sprintf("Job with id %s not found", requestURI.JobID))
		return
	}

	context.JSONResponse(200, job.ToDTO(jobStatusMessage(job.Status)))
}

// GetList
// @Summary Get list of jobs
// @Description Get list of jobs
// @Tags Job
// @Accept  json
// @Produce  json
// @Param all query bool false "Get all jobs"
// @Param userId query string false "User ID"
// @Param type query string false "Job type"
// @Param status query string false "Job status"
// @Success 200 {array} JobRead
// @Router /api/v1/job [get]
func GetList(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.JobList
	if err := context.GinContext.BindQuery(&requestQuery); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	jobs, err := job_service.GetMany(requestQuery.All, requestQuery.UserID, requestQuery.Type, requestQuery.Status, auth, &requestQuery.Pagination)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if jobs == nil {
		context.JSONResponse(200, make([]body.JobRead, 0))
		return
	}

	var jobsDTO []body.JobRead
	for _, job := range jobs {
		jobsDTO = append(jobsDTO, job.ToDTO(jobStatusMessage(job.Status)))
	}

	context.JSONResponse(200, jobsDTO)
}

func Update(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.JobUpdate
	if err := context.GinContext.BindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	var request body.JobUpdate
	if err := context.GinContext.BindJSON(&request); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	if !auth.IsAdmin {
		context.ErrorResponse(http.StatusForbidden, status_codes.Error, fmt.Sprintf("Only admin can update job"))
		return
	}

	exists, err := job_service.Exists(requestURI.JobID, auth)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to check job existence: %s", err))
		return
	}

	if !exists {
		context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, fmt.Sprintf("Job with id %s not found", requestURI.JobID))
		return
	}

	err = job_service.Update(requestURI.JobID, &request, auth)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to update job: %s", err))
		return
	}

	updatedJob, err := job_service.GetByID(requestURI.JobID, auth)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get updated job: %s", err))
		return
	}

	if updatedJob == nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Job not found after update"))
		return
	}

	context.JSONResponse(200, updatedJob.ToDTO(jobStatusMessage(updatedJob.Status)))
}

func jobStatusMessage(status string) string {
	switch status {
	case jobModel.StatusPending:
		return status_codes.GetMsg(status_codes.JobPending)
	case jobModel.StatusRunning:
		return status_codes.GetMsg(status_codes.JobRunning)
	case jobModel.StatusCompleted:
		return status_codes.GetMsg(status_codes.JobFinished)
	case jobModel.StatusFailed:
		return status_codes.GetMsg(status_codes.JobFailed)
	case jobModel.StatusTerminated:
		return status_codes.GetMsg(status_codes.JobTerminated)

		// deprecated
	case jobModel.StatusFinished:
		return status_codes.GetMsg(status_codes.JobFinished)
	default:
		return status_codes.GetMsg(status_codes.Unknown)
	}
}

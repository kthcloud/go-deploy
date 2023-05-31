package v1_job

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/uri"
	jobModel "go-deploy/models/sys/job"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
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
	context := app.NewContext(c)

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

	job, err := job_service.GetByID(auth.UserID, requestURI.JobID, auth.IsAdmin())
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

func jobStatusMessage(status string) string {
	switch status {
	case jobModel.StatusPending:
		return status_codes.GetMsg(status_codes.JobPending)
	case jobModel.StatusRunning:
		return status_codes.GetMsg(status_codes.JobRunning)
	case jobModel.StatusFinished:
		return status_codes.GetMsg(status_codes.JobFinished)
	case jobModel.StatusFailed:
		return status_codes.GetMsg(status_codes.JobFailed)
	default:
		return status_codes.GetMsg(status_codes.Unknown)
	}
}

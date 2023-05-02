package v1_job

import (
	"fmt"
	"github.com/gin-gonic/gin"
	jobModel "go-deploy/models/job"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/validator"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/job_service"
	"net/http"
)

func Get(c *gin.Context) {
	context := app.NewContext(c)

	rules := validator.MapData{
		"jobId": []string{"required", "uuid_v4"},
	}

	validationErrors := context.ValidateParams(&rules)

	if len(validationErrors) > 0 {
		context.ResponseValidationError(validationErrors)
		return
	}

	token, err := context.GetKeycloakToken()
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
	}
	jobID := context.GinContext.Param("jobId")
	userID := token.Sub
	isAdmin := v1.IsAdmin(&context)

	job, err := job_service.GetByID(userID, jobID, isAdmin)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if job == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.Error, fmt.Sprintf("Job with id %s not found", jobID))
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

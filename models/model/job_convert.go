package model

import (
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/utils"
)

// ToDTO converts a Job to a body.JobRead DTO.
func (job *Job) ToDTO(statusMessage string) body.JobRead {
	var lastError *string
	if len(job.ErrorLogs) > 0 {
		lastError = &job.ErrorLogs[len(job.ErrorLogs)-1]
	}

	return body.JobRead{
		ID:         job.ID,
		UserID:     job.UserID,
		Type:       job.Type,
		Status:     statusMessage,
		LastError:  lastError,
		CreatedAt:  job.CreatedAt,
		LastRunAt:  utils.NonZeroOrNil(job.LastRunAt),
		FinishedAt: utils.NonZeroOrNil(job.FinishedAt),
		RunAfter:   job.RunAfter,
	}
}

func (params *JobUpdateParams) FromDTO(jobUpdateDTO *body.JobUpdate) {
	params.Status = jobUpdateDTO.Status
}

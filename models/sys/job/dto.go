package job

import "go-deploy/models/dto/body"

func (job *Job) ToDTO(statusMessage string) body.JobRead {
	var lastError *string
	if len(job.ErrorLogs) > 0 {
		lastError = &job.ErrorLogs[len(job.ErrorLogs)-1]
	}

	return body.JobRead{
		ID:        job.ID,
		UserID:    job.UserID,
		Type:      job.Type,
		Status:    statusMessage,
		LastError: lastError,
	}
}

func (params *UpdateParams) FromDTO(jobUpdateDTO *body.JobUpdate) {
	params.Status = jobUpdateDTO.Status
}

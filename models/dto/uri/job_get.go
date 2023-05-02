package uri

type JobGet struct {
	JobID string `uri:"jobId" binding:"required,uuid4"`
}

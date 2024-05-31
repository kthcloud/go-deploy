package uri

type JobGet struct {
	JobID string `uri:"jobId" binding:"required,uuid4"`
}

type JobUpdate struct {
	JobID string `uri:"jobId" binding:"required,uuid4"`
}

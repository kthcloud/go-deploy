package job

import "go-deploy/models/dto/body"

func (params *UpdateParams) FromDTO(jobUpdateDTO *body.JobUpdate) {
	params.Status = jobUpdateDTO.Status
}

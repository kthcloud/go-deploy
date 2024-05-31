package model

import (
	"go-deploy/dto/v2/body"
)

type ResourceMigrationCreateParams struct {
	Type         string      `bson:"type"`
	ResourceType string      `bson:"resourceType"`
	ResourceID   string      `bson:"resourceID"`
	Params       interface{} `bson:"params"`
}

type ResourceMigrationUpdateParams struct {
	Status string `bson:"status"`
}

func (r ResourceMigrationUpdateParams) FromDTO(dto *body.ResourceMigrationUpdate) *ResourceMigrationUpdateParams {
	r.Status = dto.Status
	return &r
}

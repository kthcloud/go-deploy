package model

import (
	body "github.com/kthcloud/go-deploy/api/openapi/v2"
)

type ResourceMigrationCreateParams struct {
	Type         string `bson:"type"`
	ResourceType string `bson:"resourceType"`
	ResourceID   string `bson:"resourceID"`
	Params       any    `bson:"params"`
}

type ResourceMigrationUpdateParams struct {
	Status string `bson:"status"`
}

func (r ResourceMigrationUpdateParams) FromDTO(dto *body.ResourceMigrationUpdate) *ResourceMigrationUpdateParams {
	return dtoConverter.ResourceMigrationUpdateFromDTO(dto)
}

package job

import (
	"go-deploy/models"
	"go-deploy/models/sys/resource"
)

type Client struct {
	resource.ResourceClient[Job]
}

func New() *Client {
	return &Client{
		ResourceClient: resource.ResourceClient[Job]{
			Collection:     models.JobCollection,
			IncludeDeleted: false,
		},
	}
}

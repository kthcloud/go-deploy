package user

import (
	"go-deploy/models"
	"go-deploy/models/sys/resource"
)

type Client struct {
	resource.ResourceClient[User]
}

func New() *Client {
	return &Client{
		ResourceClient: resource.ResourceClient[User]{
			Collection:     models.UserCollection,
			IncludeDeleted: false,
		},
	}
}

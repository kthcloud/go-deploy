package notification_service

import (
	"go-deploy/service"
)

// GetOpts is used to pass options to the Get method
type GetOpts struct {
}

// ListOpts is used to pass options to the List method
type ListOpts struct {
	Pagination *service.Pagination
	UserID     *string
}

package user_service

import (
	"go-deploy/service"
)

// GetUserOpts is used to pass options to the Get method
type GetUserOpts struct {
}

// ListUsersOpts is used to pass options to the List method
type ListUsersOpts struct {
	Pagination *service.Pagination
	Search     *string
}

type DiscoverUsersOpts struct {
	Pagination *service.Pagination
	Search     *string
}

// GetTeamOpts is used to pass options to the Get method
type GetTeamOpts struct {
}

// ListTeamsOpts is used to pass options to the List method
type ListTeamsOpts struct {
	Pagination *service.Pagination
	UserID     string
	ResourceID string
}

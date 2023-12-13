package user_service

import (
	"go-deploy/service"
)

type GetTeamOpts struct {
}

type ListTeamsOpts struct {
	Pagination *service.Pagination
	UserID     string
}

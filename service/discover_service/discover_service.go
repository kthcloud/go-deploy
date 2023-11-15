package discover_service

import (
	"go-deploy/models/sys/discover"
	"go-deploy/pkg/config"
)

func Discover() (*discover.Discover, error) {
	return &discover.Discover{
		Version: config.Config.Version,
		Roles:   config.Config.Roles,
	}, nil
}

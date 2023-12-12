package client

import (
	storageManagerModel "go-deploy/models/sys/storage_manager"
	"go-deploy/service"
)

type Context struct {
	smStore map[string]*storageManagerModel.StorageManager

	Auth *service.AuthInfo
}

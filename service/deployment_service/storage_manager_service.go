package deployment_service

import (
	"fmt"
	"go-deploy/models/sys/deployment/storage_manager"
	"go-deploy/service/deployment_service/k8s_service"
)

func CreateStorageManager(id string, params *storage_manager.CreateParams) error {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to create storage manager. details: %s", err)
	}

	if params == nil {
		return makeErr(fmt.Errorf("params is nil"))
	}

	id, err := storage_manager.CreateStorageManager(id, params.UserID, params.Zone)
	if err != nil {
		return makeErr(err)
	}

	err = k8s_service.CreateStorageManager(id, params)
	if err != nil {
		return makeErr(err)
	}

	return nil
}

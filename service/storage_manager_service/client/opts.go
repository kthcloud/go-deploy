package client

import (
	"go-deploy/service"
)

// Opts is used to specify which resources to get.
// For example, if you want to get only the deployment, you can use OptsOnlyDeployment.
// If you want to get only the client, you can use OptsOnlyClient.
// If you want to get both the deployment and the client, you can use OptsAll.
type Opts struct {
	StorageManager bool
	Client         bool
	Generator      bool
}

var (
	OptsAll = &Opts{
		StorageManager: true,
		Client:         true,
		Generator:      true,
	}
	OptsNoStorageManager = &Opts{
		StorageManager: false,
		Client:         true,
		Generator:      true,
	}
	OptsNoGenerator = &Opts{
		StorageManager: true,
		Client:         true,
		Generator:      false,
	}
	OptsOnlyStorageManager = &Opts{
		StorageManager: true,
		Client:         false,
		Generator:      false,
	}
	OptsOnlyClient = &Opts{
		StorageManager: false,
		Client:         true,
		Generator:      false,
	}
)

// ListOptions is used to specify the options when listing storage managers.
type ListOptions struct {
	Pagination *service.Pagination
}

// GetOptions is used to specify the options when getting a storage manager.
type GetOptions struct {
}

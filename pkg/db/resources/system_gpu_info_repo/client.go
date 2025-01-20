package system_gpu_info_repo

import (
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/pkg/db"
	"github.com/kthcloud/go-deploy/pkg/db/resources/base_clients"
)

// Client is the client for the system GPU info model.
// It is used to query the database for data polled at a certain timestamp.
type Client struct {
	base_clients.TimestampedResourceClient[body.TimestampedSystemGpuInfo]
}

// New returns a new event client.
func New(n ...int) *Client {
	maxDocs := 1
	if len(n) > 0 {
		maxDocs = n[0]
	}

	return &Client{
		TimestampedResourceClient: base_clients.TimestampedResourceClient[body.TimestampedSystemGpuInfo]{
			Collection:   db.DB.GetCollection("systemGpuInfo"),
			MaxDocuments: maxDocs,
		},
	}
}

package vms

import (
	"github.com/kthcloud/go-deploy/service/clients"
	"github.com/kthcloud/go-deploy/service/core"
	"github.com/kthcloud/go-deploy/service/v2/api"
	"github.com/kthcloud/go-deploy/service/v2/vms/client"
	"github.com/kthcloud/go-deploy/service/v2/vms/gpu_groups"
	"github.com/kthcloud/go-deploy/service/v2/vms/gpu_leases"
	"github.com/kthcloud/go-deploy/service/v2/vms/k8s_service"
	"github.com/kthcloud/go-deploy/service/v2/vms/snapshots"
)

// Client is the client for the Deployment service.
// It is used as a wrapper around the Client.
type Client struct {
	V2 clients.V2

	client.BaseClient[Client]
}

// New creates a new VM service client.
func New(v2 clients.V2, cache ...*core.Cache) *Client {
	var ca *core.Cache
	if len(cache) > 0 {
		ca = cache[0]
	} else {
		ca = core.NewCache()
	}

	c := &Client{V2: v2, BaseClient: client.NewBaseClient[Client](ca)}
	c.BaseClient.SetParent(c)
	return c
}

// GpuLeases returns the client for the GPU Leases service.
func (c *Client) GpuLeases() api.GpuLeases {
	return gpu_leases.New(c.V2, c.Cache)
}

// GpuGroups returns the client for the GPU Groups service.
func (c *Client) GpuGroups() api.GpuGroups {
	return gpu_groups.New(c.V2, c.Cache)
}

// Snapshots returns the client for the Snapshots service.
func (c *Client) Snapshots() api.Snapshots {
	return snapshots.New(c.V2, c.Cache)
}

// K8s returns the client for the K8s service.
func (c *Client) K8s() *k8s_service.Client {
	return k8s_service.New(c.Cache)
}

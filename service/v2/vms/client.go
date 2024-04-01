package vms

import (
	"go-deploy/service/clients"
	"go-deploy/service/core"
	"go-deploy/service/v2/api"
	"go-deploy/service/v2/vms/client"
	"go-deploy/service/v2/vms/gpu_groups"
	"go-deploy/service/v2/vms/gpu_leases"
	"go-deploy/service/v2/vms/gpus"
	"go-deploy/service/v2/vms/k8s_service"
	"go-deploy/service/v2/vms/snapshots"
)

// Client is the client for the Deployment service.
// It is used as a wrapper around the Client.
type Client struct {
	V1 clients.V1
	V2 clients.V2

	client.BaseClient[Client]
}

// New creates a new VM service client.
func New(v1 clients.V1, v2 clients.V2, cache ...*core.Cache) *Client {
	var ca *core.Cache
	if len(cache) > 0 {
		ca = cache[0]
	} else {
		ca = core.NewCache()
	}

	c := &Client{V1: v1, V2: v2, BaseClient: client.NewBaseClient[Client](ca)}
	c.BaseClient.SetParent(c)
	return c
}

// GPUs returns the client for the GPUs service.
func (c *Client) GPUs() api.GPUs {
	return gpus.New(c.V1, c.V2, c.Cache)
}

// GpuLeases returns the client for the GPU Leases service.
func (c *Client) GpuLeases() api.GpuLeases {
	return gpu_leases.New(c.V1, c.V2, c.Cache)
}

// GpuGroups returns the client for the GPU Groups service.
func (c *Client) GpuGroups() api.GpuGroups {
	return gpu_groups.New(c.V1, c.V2, c.Cache)
}

// Snapshots returns the client for the Snapshots service.
func (c *Client) Snapshots() api.Snapshots {
	return snapshots.New(c.V1, c.V2, c.Cache)
}

// K8s returns the client for the K8s service.
func (c *Client) K8s() *k8s_service.Client {
	return k8s_service.New(c.Cache)
}

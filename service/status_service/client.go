package status_service

import "go-deploy/service"

// Client is the client for the Status service.
type Client struct {
	// Cache is used to cache the resources fetched inside the service.
	Cache *service.Cache

	// Auth is the authentication information for the client.
	Auth *service.AuthInfo
}

// New creates a new Status service client.
func New() *Client {
	return &Client{
		Cache: service.NewCache(),
	}
}

// WithAuth sets the auth on the context.
// This is used to perform authorization checks.
func (c *Client) WithAuth(auth *service.AuthInfo) *Client {
	c.Auth = auth
	return c
}

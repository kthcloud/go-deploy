package k8s_service

import (
	"go-deploy/models/config"
	storageManagerModel "go-deploy/models/sys/storage_manager"
	"go-deploy/pkg/subsystems/k8s"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/resources"
	"go-deploy/service/storage_manager_service/client"
	"go-deploy/utils/subsystemutils"
)

// Client is the client for the Harbor service.
// It contains a BaseClient, which is used to lazy-load and cache data.
type Client struct {
	client.BaseClient[Client]

	client    *k8s.Client
	generator *resources.K8sGenerator
}

// New creates a new Client.
// If context is not nil, it will be used to create a new BaseClient.
// Otherwise, an empty context will be created.
func New(context *client.Context) *Client {
	c := &Client{}
	c.BaseClient.SetParent(c)
	if context != nil {
		c.BaseClient.SetContext(context)
	}

	return c
}

// Get returns the deployment, client, and generator.
//
// Depending on the options specified, some return values may be nil.
// This is useful when you don't always need all the resources.
//
// The default option is OptsAll.
func (c *Client) Get(opts *client.Opts) (*storageManagerModel.StorageManager, *k8s.Client, *resources.K8sGenerator, error) {
	if opts == nil {
		opts = client.OptsAll
	}

	var sm *storageManagerModel.StorageManager
	if opts.StorageManager {
		sm = c.StorageManager()
		if sm == nil {
			return nil, nil, nil, sErrors.DeploymentNotFoundErr
		}
	}

	var kc *k8s.Client
	if opts.Client {
		kc = c.Client()
		if kc == nil {
			return nil, nil, nil, sErrors.DeploymentNotFoundErr
		}
	}

	var g *resources.K8sGenerator
	if opts.Generator {
		g = c.Generator()
		if g == nil {
			return nil, nil, nil, sErrors.DeploymentNotFoundErr
		}
	}

	return sm, kc, g, nil
}

// WithUserID sets the user id
// Overwrites the base client's user id function
// This is used to set the namespace
func (c *Client) WithUserID(userID string) *Client {
	kc := c.Client()
	if kc != nil {
		kc.Namespace = getNamespaceName(userID)
	}

	g := c.Generator()
	if g != nil {
		g = g.K8s(c.Client())
	}

	c.BaseClient.WithUserID(userID)

	return c
}

// Client returns the K8s service client.
//
// If the client does not exist, it will be created.
func (c *Client) Client() *k8s.Client {
	if c.client == nil {
		if !c.HasUserID() {
			panic("user id is empty")
		}

		c.client = withClient(c.Zone(), getNamespaceName(c.UserID()))
	}

	return c.client
}

// Generator returns the K8s generator.
//
// If the generator does not exist, it will be created.
// If creating a new generator, the current deployment and zone will be used.
// Set the deployment and zone before calling this function by using WithDeployment and WithZone.
func (c *Client) Generator() *resources.K8sGenerator {
	if c.generator == nil {
		pg := resources.PublicGenerator()

		if c.StorageManager() != nil {
			pg.WithStorageManager(c.StorageManager())
		}

		if c.Zone() != nil {
			pg.WithDeploymentZone(c.Zone())
		}

		c.generator = pg.K8s(c.Client())
	}

	return c.generator
}

// getNamespaceName returns the namespace name for the user.
func getNamespaceName(userID string) string {
	return subsystemutils.GetPrefixedName(userID)
}

// withClient returns a new K8s service client.
func withClient(zone *config.DeploymentZone, namespace string) *k8s.Client {
	c, _ := k8s.New(zone.Client, namespace)
	return c
}

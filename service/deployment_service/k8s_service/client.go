package k8s_service

import (
	"go-deploy/models/config"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/service/deployment_service/client"
	"go-deploy/service/resources"
	"go-deploy/utils/subsystemutils"
)

type Client struct {
	client.BaseClient[Client]

	client    *k8s.Client
	generator *resources.K8sGenerator
}

func New(context *client.Context) *Client {
	c := &Client{}
	c.BaseClient.SetParent(c)
	if context != nil {
		c.BaseClient.SetContext(context)
	}

	return c
}

func (c *Client) Client() *k8s.Client {
	if c.client == nil {
		if c.UserID == "" {
			panic("user id is empty")
		}

		c.client = withClient(c.Zone(), getNamespaceName(c.UserID))
	}

	return c.client
}

func (c *Client) Generator() *resources.K8sGenerator {
	if c.generator == nil {
		pg := resources.PublicGenerator()

		if c.Deployment() != nil {
			pg.WithDeployment(c.Deployment())
		}

		if c.Zone() != nil {
			pg.WithDeploymentZone(c.Zone())
		}

		c.generator = pg.K8s(c.Client())
	}

	return c.generator
}

func getNamespaceName(userID string) string {
	return subsystemutils.GetPrefixedName(userID)
}

func withClient(zone *config.DeploymentZone, namespace string) *k8s.Client {
	c, _ := k8s.New(zone.Client, namespace)
	return c
}

package cs

import (
	"fmt"
	"time"

	"go-deploy/pkg/imp/cloudstack"
)

// ClientConf is the configuration for the CloudStack client.
type ClientConf struct {
	URL    string
	ApiKey string
	Secret string

	IpAddressID string
	NetworkID   string
	ProjectID   string
	ZoneID      string
	TemplateID  string
}

// Client is a wrapper around a CloudStack client.
type Client struct {
	CsClient                *cloudstack.CloudStackClient
	RootIpAddressID         string
	RootNetworkID           string
	NetworkOfferingID       string
	ProjectID               string
	ZoneID                  string
	TemplateID              string
	CustomServiceOfferingID string

	userSshPublicKey  *string
	adminSshPublicKey *string
}

// New creates a new CloudStack wrapper client.
func New(config *ClientConf) (*Client, error) {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to create cloudstack client. details: %w", err)
	}

	csClient := cloudstack.NewAsyncClient(
		config.URL,
		config.ApiKey,
		config.Secret,
		true,
		func(c *cloudstack.CloudStackClient) {
			c.Timeout(300 * time.Second)
		},
	)

	client := Client{
		CsClient:        csClient,
		RootIpAddressID: config.IpAddressID,
		RootNetworkID:   config.NetworkID,
		ProjectID:       config.ProjectID,
		ZoneID:          config.ZoneID,
		TemplateID:      config.TemplateID,
	}

	soID, err := client.getCustomServiceOfferingID()
	if err != nil {
		return nil, makeErr(err)
	}

	client.CustomServiceOfferingID = soID

	return &client, nil
}

// WithUserSshPublicKey sets the user SSH public key.
func (client *Client) WithUserSshPublicKey(key string) *Client {
	client.userSshPublicKey = &key
	return client
}

// WithAdminSshPublicKey sets the admin SSH public key.
func (client *Client) WithAdminSshPublicKey(key string) *Client {
	client.adminSshPublicKey = &key
	return client
}

package cs

import (
	"fmt"

	"go-deploy/pkg/imp/cloudstack"
)

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

type Client struct {
	CsClient          *cloudstack.CloudStackClient
	RootIpAddressID   string
	RootNetworkID     string
	NetworkOfferingID string
	ProjectID         string
	ZoneID            string
	TemplateID        string

	userSshPublicKey  *string
	adminSshPublicKey *string
}

func New(config *ClientConf) (*Client, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to create cloudstack client. details: %w", err)
	}

	csClient := cloudstack.NewAsyncClient(
		config.URL,
		config.ApiKey,
		config.Secret,
		true,
	)

	client := Client{
		CsClient:        csClient,
		RootIpAddressID: config.IpAddressID,
		RootNetworkID:   config.NetworkID,
		ProjectID:       config.ProjectID,
		ZoneID:          config.ZoneID,
		TemplateID:      config.TemplateID,
	}

	return &client, nil
}

func (client *Client) WithUserSshPublicKey(key string) *Client {
	client.userSshPublicKey = &key
	return client
}

func (client *Client) WithAdminSshPublicKey(key string) *Client {
	client.adminSshPublicKey = &key
	return client
}

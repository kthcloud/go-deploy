package sys_api

import (
	"context"
	"fmt"
	"github.com/Nerzal/gocloak/v13"
	"net/http"
)

// Client is a client for the sys-api service.
type Client struct {
	url          string
	oauth2Client *http.Client
	jwt          *gocloak.JWT
}

// ClientConf is the configuration for creating a sys-api client.
type ClientConf struct {
	URL      string
	Username string
	Password string

	OidcProvider string
	OidcClientID string
	OidcRealm    string
}

// New creates a new sys-api client.
func New(config *ClientConf) (*Client, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create sys-api oauth2 client. details: %w", err)
	}

	kcClient := gocloak.NewClient(config.OidcProvider)

	jwt, err := kcClient.Login(context.TODO(), config.OidcClientID, "", config.OidcRealm, config.Username, config.Password)
	if err != nil {
		return nil, makeError(err)
	}

	client := &Client{
		url: config.URL,
		jwt: jwt,
	}

	return client, nil
}

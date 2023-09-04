package landing

import (
	"context"
	"fmt"
	"github.com/Nerzal/gocloak/v13"
	"net/http"
)

type Client struct {
	url          string
	oauth2Client *http.Client
	jwt          *gocloak.JWT
}

type ClientConf struct {
	URL      string
	Username string
	Password string

	OidcProvider string
	OidcClientID string
	OidcRealm    string
}

func New(config *ClientConf) (*Client, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create landing oauth2Client. details: %w", err)
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

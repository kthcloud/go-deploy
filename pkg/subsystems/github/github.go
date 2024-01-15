package github

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Client is a wrapper around the github.Client.
type Client struct {
	GitHubClient *github.Client
}

// ClientConf is the configuration for the GitHub client.
type ClientConf struct {
	Token string
}

// New creates a new GitHub wrapper client.
func New(config *ClientConf) (*Client, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to create github client. details: %w", err)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	tc := oauth2.NewClient(context.TODO(), ts)

	githubClient := github.NewClient(tc)

	client := Client{
		GitHubClient: githubClient,
	}

	return &client, nil
}

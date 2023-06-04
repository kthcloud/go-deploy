package github

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type Client struct {
	GitHubClient *github.Client
}

type ClientConf struct {
	Token string
	Code  string
}

func New(config *ClientConf) (*Client, error) {
	_ = func(err error) error {
		return fmt.Errorf("failed to create github client. details: %s", err)
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

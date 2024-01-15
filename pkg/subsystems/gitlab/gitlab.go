package gitlab

import (
	"fmt"
	"github.com/xanzy/go-gitlab"
)

// Client is a wrapper around the gitlab.Client.
type Client struct {
	GitLabClient *gitlab.Client
}

// ClientConf is the configuration for the GitLab client.
type ClientConf struct {
	URL   string
	Token string
}

// New creates a new GitLab wrapper client.
func New(config *ClientConf) (*Client, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create gitlab client. details: %w", err)
	}

	gitlabClient, err := gitlab.NewClient(config.Token, gitlab.WithBaseURL(config.URL))
	if err != nil {

		return nil, makeError(err)
	}

	client := Client{
		GitLabClient: gitlabClient,
	}

	return &client, nil
}

package gitlab

import (
	"fmt"
	"github.com/xanzy/go-gitlab"
)

type Client struct {
	GitLabClient *gitlab.Client
}

type ClientConf struct {
	URL   string
	Token string
}

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

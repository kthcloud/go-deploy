package terraform

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
)

type Client struct {
	workingDir string
	execPath   string
	externals  []External
}

type ClientConf struct {
	WorkingDir string
}

func New(config *ClientConf) (*Client, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create cloudstack client. details: %s", err)
	}

	execPath, err := setupTerraform()
	if err != nil {
		return nil, makeError(err)
	}

	client := Client{
		execPath:   execPath,
		workingDir: config.WorkingDir,
	}

	return &client, nil
}

func setupTerraform() (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to setup terraform. details: %s", err)
	}

	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("1.0.6")),
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		return "", makeError(err)
	}

	return execPath, nil
}

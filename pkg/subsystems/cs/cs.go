package cs

import (
	"context"
	"fmt"
	"github.com/apache/cloudstack-go/v2/cloudstack"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	"os"
)

type Client struct {
	CSClient  *cloudstack.CloudStackClient
	ZoneID    string
	Terraform *tfexec.Terraform
}

type ClientConf struct {
	ApiUrl       string
	ApiKey       string
	SecretKey    string
	ZoneID       string
	TerraformDir string
}

func New(config *ClientConf) (*Client, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create cloudstack client. details: %s", err)
	}

	csClient := cloudstack.NewAsyncClient(
		config.ApiUrl,
		config.ApiKey,
		config.SecretKey,
		true,
	)

	terraform, err := setupTerraform(config.TerraformDir, config.ApiUrl, config.ApiKey, config.SecretKey)
	if err != nil {
		return nil, makeError(err)
	}

	client := Client{
		CSClient:  csClient,
		ZoneID:    config.ZoneID,
		Terraform: terraform,
	}

	return &client, nil
}

func setupTerraform(workingDir string, apiUrl, apiKey, secretKey string) (*tfexec.Terraform, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to setup terraform. details: %s", err)
	}

	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("1.0.6")),
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		return nil, makeError(err)
	}

	terraformClient, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		return nil, makeError(err)
	}

	err = os.Setenv("TF_VAR_cloudstack_api_url", apiUrl)
	if err != nil {
		return nil, makeError(err)
	}
	err = os.Setenv("TF_VAR_cloudstack_api_key", apiKey)
	if err != nil {
		return nil, makeError(err)
	}
	err = os.Setenv("TF_VAR_cloudstack_secret_key", secretKey)
	if err != nil {
		return nil, makeError(err)
	}

	err = terraformClient.Init(
		context.Background(),
		tfexec.Upgrade(true),
	)
	if err != nil {
		return nil, makeError(err)
	}

	return terraformClient, nil
}

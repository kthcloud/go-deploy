package cs

import (
	"context"
	"fmt"
	"github.com/apache/cloudstack-go/v2/cloudstack"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/terraform"
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

func GetTerraform() *terraform.External {
	//vars, err := filepath.Abs("terraform/subsystems/cs/cs_vars.tf")
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//
	//vmSmall, err := filepath.Abs("terraform/subsystems/cs/cs_vm_small.tf")
	//if err != nil {
	//	log.Fatalln(err)
	//}

	return &terraform.External{
		Envs: map[string]string{
			"TF_VAR_cloudstack_api_url":    conf.Env.CS.Url,
			"TF_VAR_cloudstack_api_key":    conf.Env.CS.Key,
			"TF_VAR_cloudstack_secret_key": conf.Env.CS.Secret,
		},
		//ScriptPaths: []string{vars, vmSmall},
		Provider: terraform.Provider{
			Name:    "cloudstack",
			Source:  "cloudstack/cloudstack",
			Version: "0.4.0",
		},
	}
}

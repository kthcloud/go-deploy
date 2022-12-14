package terraform

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-exec/tfexec"
	"go-deploy/pkg/conf"
	"log"
	"os"
	"path"
	"path/filepath"
)

func (client *Client) Apply() error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to apply terraform script. details: %s", err)
	}

	terraformClient, err := tfexec.NewTerraform(client.workingDir, client.execPath)
	if err != nil {
		return makeError(err)
	}

	for _, externalGenerator := range client.externalGenerators {

		external := External{
			envs:        map[string]string{},
			scriptPaths: []string{},
		}
		err = externalGenerator(&external)
		if err != nil {
			log.Println("failed to apply external generator. details:", err)
			continue
		}

		err = client.copyScripts(external.scriptPaths)
		if err != nil {
			log.Println("failed to apply external generator. details:", err)
			continue
		}

		err = applyEnvs(external.envs)
		if err != nil {
			return makeError(err)
		}
	}

	configStr := fmt.Sprintf(
		"conn_str=postgres://%s:%s@%s/terraform_backend?sslmode=disable",
		conf.Env.Terraform.Username,
		conf.Env.Terraform.Password,
		conf.Env.Terraform.Url,
	)

	err = terraformClient.Init(context.Background(), tfexec.Upgrade(true), tfexec.BackendConfig(configStr), tfexec.Reconfigure(true))
	if err != nil {
		return makeError(err)
	}

	err = terraformClient.Apply(context.TODO())
	if err != nil {
		return makeError(err)
	}

	return nil
}

func (client *Client) copyScripts(scriptPaths []string) error {
	for _, scriptPath := range scriptPaths {
		err := client.copyScript(scriptPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (client *Client) copyScript(from string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to copy script from %s to working dir. details: %s", from, err)
	}

	filename := filepath.Base(from)
	if filename == "." {
		log.Println("trying to copy a path with no filename:", from)
		return nil
	}

	workingDirPath := path.Join(client.workingDir, filename)
	_, err := os.Stat(workingDirPath)

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return makeError(err)
	}

	if err != nil && errors.Is(err, os.ErrNotExist) {
		copyErr := Copy(from, workingDirPath)
		if copyErr != nil {
			return makeError(copyErr)
		}
	}

	return nil
}

func applyEnvs(envs map[string]string) error {
	makeError := func(err error, env string) error {
		return fmt.Errorf("failed to set env %s. details: %s", env, err)
	}

	for key, val := range envs {
		err := os.Setenv(key, val)
		if err != nil {
			return makeError(err, key)
		}
	}

	return nil
}

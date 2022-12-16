package npm

import (
	"fmt"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/npm/models"
	"go-deploy/pkg/terraform"
	"go-deploy/utils/requestutils"
)

type Client struct {
	apiUrl string
	token  string
}

type ClientConf struct {
	ApiUrl   string
	Username string
	Password string
}

func New(config *ClientConf) (*Client, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create npm client. details: %s", err)
	}

	client := Client{
		apiUrl: config.ApiUrl,
	}

	token, err := client.createToken(config.Username, config.Password)
	if err != nil {
		return nil, makeError(err)
	}

	client.token = token

	return &client, nil
}

func (client *Client) createToken(username, password string) (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create token. details: %s", err)
	}

	tokenReq := models.TokenReq{Identity: username, Secret: password}
	res, err := client.doJSONRequestUnauthorized("POST", "/tokens", tokenReq)

	// check if good request
	if !requestutils.IsGoodStatusCode(res.StatusCode) {
		return "", makeApiError(res.Body, makeError)
	}

	var token models.Token
	err = requestutils.ParseBody(res.Body, &token)
	if err != nil {
		return "", makeError(err)
	}

	return token.Token, nil

}

func GetTerraform() *terraform.External {
	//vars, err := filepath.Abs("terraform/subsystems/npm/npm_vars.tf")
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//
	//generatedProxyHost, err := terraform.CreateTfResource(
	//	"nginx_proxy_manager_proxy_host",
	//	"web",
	//	map[string]cty.Value{
	//		"domain_names": cty.ListVal([]cty.Value{
	//			cty.StringVal("test.dev.kthcloud.com"),
	//		}),
	//	})

	return &terraform.External{
		Envs: map[string]string{
			"TF_VAR_NPM_api_url":  conf.Env.NPM.Url,
			"TF_VAR_NPM_username": conf.Env.NPM.Identity,
			"TF_VAR_NPM_password": conf.Env.NPM.Secret,
		},
		//ScriptPaths: []string{vars},
		//Provider: terraform.Provider{
		//	Name:    "nginx-proxy-manager",
		//	Source:  "saffronjam/nginx-proxy-manager",
		//	Version: "0.2.3",
		//},
		//Scripts: map[string]*hclwrite.File{"npm_proxy_host.tf": generatedProxyHost},
	}
}

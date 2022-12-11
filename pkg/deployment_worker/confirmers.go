package deployment_worker

import (
	"fmt"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/npm"
)

func getFQDN(name string) string {
	return fmt.Sprintf("%s.%s", name, conf.Env.ParentDomain)
}

func NPMCreated(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if npm setup is created for deployment %s. details: %s", name, err)
	}

	client, err := npm.New(&npm.ClientConf{
		ApiUrl:   conf.Env.Npm.Url,
		Username: conf.Env.Npm.Identity,
		Password: conf.Env.Npm.Secret,
	})
	if err != nil {
		return false, makeError(err)
	}

	proxyHostCreated, err := client.ProxyHostCreated(getFQDN(name))
	if err != nil {
		return false, makeError(err)
	}

	return proxyHostCreated, nil
}

func NPMDeleted(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if npm setup is deleted for deployment %s. details: %s", name, err)
	}

	client, err := npm.New(&npm.ClientConf{
		ApiUrl:   conf.Env.Npm.Url,
		Username: conf.Env.Npm.Identity,
		Password: conf.Env.Npm.Secret,
	})
	if err != nil {
		return false, makeError(err)
	}

	proxyHostDeleted, err := client.ProxyHostDeleted(getFQDN(name))
	if err != nil {
		return false, makeError(err)
	}

	return proxyHostDeleted, nil
}

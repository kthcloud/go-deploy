package subsystem_service

import (
	"fmt"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/npm"
	"go-deploy/utils/subsystemutils"
	"log"
)

func getFQDN(name string) string {
	return fmt.Sprintf("%s.%s", name, conf.Env.ParentDomain)
}

func CreateNPM(name string) error {
	log.Println("setup npm for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup npm for deployment %s. details: %s", name, err)
	}

	client, err := npm.New(&npm.ClientConf{
		ApiUrl:   conf.Env.NPM.Url,
		Username: conf.Env.NPM.Identity,
		Password: conf.Env.NPM.Secret,
	})
	if err != nil {
		return makeError(err)
	}

	certificateID, err := client.GetWildcardCertificateID(conf.Env.ParentDomain)
	if err != nil {
		return makeError(err)
	}

	forwardHost := fmt.Sprintf("%s.%s.svc.cluster.local", name, subsystemutils.GetPrefixedName(name))
	err = client.CreateProxyHost(getFQDN(name), forwardHost, conf.Env.AppPort, certificateID)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func DeleteNPM(name string) error {
	log.Println("deleting npm for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup npm for deployment %s. details: %s", name, err)
	}

	client, err := npm.New(&npm.ClientConf{
		ApiUrl:   conf.Env.NPM.Url,
		Username: conf.Env.NPM.Identity,
		Password: conf.Env.NPM.Secret,
	})
	if err != nil {
		return makeError(err)
	}

	err = client.DeleteProxyHost(getFQDN(name))
	if err != nil {
		return makeError(err)
	}

	return nil
}

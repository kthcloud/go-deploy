package deployment_service

import (
	"fmt"
	deploymentModel "go-deploy/models/deployment"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/npm"
	npmModels "go-deploy/pkg/subsystems/npm/models"
	"go-deploy/utils/subsystemutils"
	"go.mongodb.org/mongo-driver/bson"
	"log"
)

func getFQDN(name string) string {
	return fmt.Sprintf("%s.%s", name, conf.Env.ParentDomain)
}

func createProxyHostPublicBody(name string, certificateId int) *npmModels.ProxyHostPublic {
	forwardHost := fmt.Sprintf("%s.%s.svc.cluster.local", name, subsystemutils.GetPrefixedName(name))

	return &npmModels.ProxyHostPublic{
		ID:                    0,
		DomainNames:           []string{getFQDN(name)},
		ForwardHost:           forwardHost,
		ForwardPort:           conf.Env.AppPort,
		CertificateID:         certificateId,
		AllowWebsocketUpgrade: false,
		ForwardScheme:         "http",
		Enabled:               true,
		Locations:             []npmModels.Location{},
	}
}

func CreateNPM(name string) error {
	log.Println("setting up npm for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup npm for v1_deployment %s. details: %s", name, err)
	}

	client, err := npm.New(&npm.ClientConf{
		ApiUrl:   conf.Env.NPM.Url,
		Username: conf.Env.NPM.Identity,
		Password: conf.Env.NPM.Secret,
	})
	if err != nil {
		return makeError(err)
	}

	deployment, err := deploymentModel.GetDeploymentByName(name)

	if deployment.Subsystems.Npm.ProxyHost.ID == 0 {
		certificateID, err := client.GetWildcardCertificateID(conf.Env.ParentDomain)
		if err != nil {
			return makeError(err)
		}

		id, err := client.CreateProxyHost(createProxyHostPublicBody(name, certificateID))
		if err != nil {
			return err
		}
		deployment.Subsystems.Npm.ProxyHost.ID = id

		proxyHost, err := client.ReadProxyHost(deployment.Subsystems.Npm.ProxyHost.ID)
		deployment.Subsystems.Npm.ProxyHost = *proxyHost

		err = deploymentModel.UpdateSubsystemByName(name, "npm", "proxyHost", *proxyHost)
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func DeleteNPM(name string) error {
	log.Println("deleting npm for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup npm for v1_deployment %s. details: %s", name, err)
	}

	client, err := npm.New(&npm.ClientConf{
		ApiUrl:   conf.Env.NPM.Url,
		Username: conf.Env.NPM.Identity,
		Password: conf.Env.NPM.Secret,
	})
	if err != nil {
		return makeError(err)
	}

	deployment, err := deploymentModel.GetDeploymentByName(name)

	if deployment.Subsystems.Npm.ProxyHost.ID == 0 {
		return nil
	}

	err = client.DeleteProxyHost(deployment.Subsystems.Npm.ProxyHost.ID)
	if err != nil {
		return makeError(err)
	}

	err = deploymentModel.UpdateDeploymentByName(name, bson.D{{"subsystems.npm.proxyHost", npmModels.ProxyHostPublic{}}})
	if err != nil {
		return makeError(err)
	}

	return nil
}

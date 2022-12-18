package subsystem_service

import (
	"fmt"
	"go-deploy/models"
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

	deployment, err := models.GetDeploymentByName(name)

	public := createProxyHostPublicBody(name, certificateID)
	var proxyHost *npmModels.ProxyHostPublic

	if deployment.Subsytems.Npm.Public.ID == 0 {
		id, err := client.CreateProxyHost(public)
		if err != nil {
			return err
		}
		deployment.Subsytems.Npm.Public.ID = id
	} else {
		public.ID = deployment.Subsytems.Npm.Public.ID
		err = client.UpdateProxyHost(public)
		if err != nil {
			return makeError(err)
		}
	}

	proxyHost, err = client.ReadProxyHost(deployment.Subsytems.Npm.Public.ID)
	deployment.Subsytems.Npm.Public = *proxyHost

	err = models.UpdateDeployment(deployment.ID, bson.D{{"subsystems.npm.public", *proxyHost}})
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

	deployment, err := models.GetDeploymentByName(name)

	if deployment.Subsytems.Npm.Public.ID == 0 {
		return nil
	}

	err = client.DeleteProxyHost(deployment.Subsytems.Npm.Public.ID)
	if err != nil {
		return makeError(err)
	}

	err = models.UpdateDeploymentByName(name, bson.D{{"subsystems.npm.public", npmModels.ProxyHostPublic{}}})
	if err != nil {
		return makeError(err)
	}

	return nil
}

package test

import (
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/npm"
	"go-deploy/pkg/subsystems/npm/models"
	"testing"
)

func TestProxyHost(t *testing.T) {
	setup(t)

	client, err := npm.New(&npm.ClientConf{
		ApiUrl:   conf.Env.NPM.Url,
		Username: conf.Env.NPM.Identity,
		Password: conf.Env.NPM.Secret,
	})

	if err != nil {
		t.Fatalf(err.Error())
	}

	public := &models.ProxyHostPublic{
		DomainNames:           []string{"test.test"},
		ForwardHost:           "1.1.1.1",
		ForwardPort:           1111,
		CertificateID:         0,
		AllowWebsocketUpgrade: true,
		ForwardScheme:         "http",
		Enabled:               true,
		Locations:             []models.Location{},
	}

	id, err := client.CreateProxyHost(public)
	if err != nil {
		t.Fatalf(err.Error())
	}

	public.ID = id

	if public.ID == 0 {
		t.Fatalf("no proxy host id received from client")
	}

	err = client.DeleteProxyHost(public.ID)
	if err != nil {
		t.Fatalf(err.Error())
	}

	deletedProxyHost, err := client.ReadProxyHost(public.ID)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if deletedProxyHost != nil {
		t.Fatalf("failed to delete proxy host")
	}
}

func TestProxyHostUpdate(t *testing.T) {
	setup(t)

	client, err := npm.New(&npm.ClientConf{
		ApiUrl:   conf.Env.NPM.Url,
		Username: conf.Env.NPM.Identity,
		Password: conf.Env.NPM.Secret,
	})

	if err != nil {
		t.Fatalf(err.Error())
	}

	public := &models.ProxyHostPublic{
		DomainNames:           []string{"test.test"},
		ForwardHost:           "1.1.1.1",
		ForwardPort:           1111,
		CertificateID:         0,
		AllowWebsocketUpgrade: true,
		ForwardScheme:         "http",
		Enabled:               true,
		Locations:             []models.Location{},
	}

	id, err := client.CreateProxyHost(public)
	if err != nil {
		t.Fatalf(err.Error())
	}

	public.ID = id

	if public.ID == 0 {
		t.Fatalf("no proxy host id received from client")
	}

	public.ForwardPort = 2222

	err = client.UpdateProxyHost(public)
	if err != nil {
		t.Fatalf(err.Error())
	}

	updatedProxyHost, err := client.ReadProxyHost(public.ID)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if updatedProxyHost.ForwardPort != 2222 {
		t.Fatalf("failed to update proxy host forward port 1111 -> 2222")
	}

	err = client.DeleteProxyHost(public.ID)
	if err != nil {
		t.Fatalf(err.Error())
	}

	deletedProxyHost, err := client.ReadProxyHost(public.ID)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if deletedProxyHost != nil {
		t.Fatalf("failed to delete proxy host")
	}
}

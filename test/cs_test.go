package test

import (
	"github.com/google/uuid"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/cs"
	"go-deploy/pkg/subsystems/cs/models"
	"testing"
)

func TestCreateVM(t *testing.T) {
	setup(t)

	client, err := cs.New(&cs.ClientConf{
		ApiUrl:    conf.Env.CS.Url,
		ApiKey:    conf.Env.CS.Key,
		SecretKey: conf.Env.CS.Secret,
	})

	if err != nil {
		t.Fatalf(err.Error())
	}

	public := &models.VmPublic{
		Name:              "acc-test-" + uuid.New().String(),
		ServiceOfferingID: "8da28b4d-5fec-4a44-aee7-fb0c5c8265a9", // Small HA
		TemplateID:        "e1a0479c-76a2-44da-8b38-a3a3fa316287", // Ubuntu Server
		NetworkID:         "4a065a52-f290-4d2e-aeb4-6f48d3bd9bfe", // deploy
		ZoneID:            "3a74db73-6058-4520-8d8c-ab7d9b7955c8", // Flemingsberg
		ProjectID:         "d1ba382b-e310-445b-a54b-c4e773662af3", // deploy
	}

	id, err := client.CreateVM(public)

	if err != nil {
		t.Fatalf(err.Error())
	}

	public.ID = id

	if len(public.ID) == 0 {
		t.Fatalf("no vm id received from client")
	}

	err = client.DeleteVM(public.ID)
	if err != nil {
		t.Fatalf(err.Error())
	}

	deletedProject, err := client.ReadVM(public.ID)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if deletedProject != nil {
		t.Fatalf("failed to delete vm")
	}
}

func TestUpdateVM(t *testing.T) {
	setup(t)

	client, err := cs.New(&cs.ClientConf{
		ApiUrl:    conf.Env.CS.Url,
		ApiKey:    conf.Env.CS.Key,
		SecretKey: conf.Env.CS.Secret,
	})

	if err != nil {
		t.Fatalf(err.Error())
	}

	public := &models.VmPublic{
		Name:              "acc-test-" + uuid.New().String(),
		ServiceOfferingID: "8da28b4d-5fec-4a44-aee7-fb0c5c8265a9", // Small HA
		TemplateID:        "e1a0479c-76a2-44da-8b38-a3a3fa316287", // Ubuntu Server
		NetworkID:         "4a065a52-f290-4d2e-aeb4-6f48d3bd9bfe", // deploy
		ZoneID:            "3a74db73-6058-4520-8d8c-ab7d9b7955c8", // Flemingsberg
		ProjectID:         "d1ba382b-e310-445b-a54b-c4e773662af3", // deploy
	}

	id, err := client.CreateVM(public)

	if err != nil {
		t.Fatalf(err.Error())
	}

	public.ID = id

	if len(public.ID) == 0 {
		t.Fatalf("no vm id received from client")
	}

	public.Name = public.Name + "-increased"
	public.ExtraConfig = "some gpu config"

	err = client.UpdateVM(public)
	if err != nil {
		t.Fatalf(err.Error())
	}

	updated, err := client.ReadVM(public.ID)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if updated.Name != public.Name {
		t.Fatalf("failed to update vm name")
	}

	if updated.ExtraConfig != public.ExtraConfig {
		t.Fatalf("failed to update vm extra config")
	}

	err = client.DeleteVM(public.ID)
	if err != nil {
		t.Fatalf(err.Error())
	}

	deleted, err := client.ReadVM(public.ID)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if deleted != nil {
		t.Fatalf("failed to delete vm")
	}
}

package test

import (
	"github.com/google/uuid"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/cs"
	"go-deploy/pkg/subsystems/cs/models"
	"testing"
)

func TestCreateVM(t *testing.T) {
	sshPublicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDbFXrLEF2PYNodfwNfGe+4qM3FeZ/FxcfYLZwxStKVW/eTgYn3Y0DQSti86mA+Jrzkx2aSvHDPPJEQUUTiZUMwTiJlR4ud3FBDYZXQVsNhfJO5zduqLpEEdjtFMP/Y3jGpoh+Eq8U08yWRfs1sKay/THS5MoKIprFVU+yIgHsxNcrU2hymTnt+A43dxKHXd4aZXhfW5qHwBZzNBggRXPFb6RpABx2qk9dQGGHFrGp5p0cC2sekXNFg7lV8PEx8pspu+Kf5mSBd1aphRde8ATR61zEDbAGKi1wbGHhrrZ/dAigcSB5YNgllTg5l09CwtjWBFHGY1oxwb+F3foXH19dDIlkB7wsyoT/XD7NMOfNyqDYLlOrVVMPfEdNkBXdCScK8Q3rrT/LL/7fefo/OirDnCvL3dxEA/9ay0hVEHyef6E++tiO9DU/HBVAY6NYjYQCZZ92rqVPzM94ppBU4XocxzAQ7zL+pFABbZkYtXTH8VaE4A1MTgRXvte1CmzeFPQs= emil@thinkpad"

	setup(t)

	client, err := cs.New(&cs.ClientConf{
		URL:         conf.Env.CS.URL,
		ApiKey:      conf.Env.CS.ApiKey,
		Secret:      conf.Env.CS.Secret,
		IpAddressID: conf.Env.CS.IpAddressID,
		NetworkID:   conf.Env.CS.NetworkID,
		ProjectID:   conf.Env.CS.ProjectID,
		ZoneID:      conf.Env.CS.ZoneID,
	})

	if err != nil {
		t.Fatalf(err.Error())
	}

	public := &models.VmPublic{
		Name:              "acc-test-" + uuid.New().String(),
		ServiceOfferingID: "8da28b4d-5fec-4a44-aee7-fb0c5c8265a9", // Small HA
		TemplateID:        "fb6b6b11-6196-42d9-a12d-038bdeecb6f6", // Ubuntu Server
	}

	id, err := client.CreateVM(public, sshPublicKey, conf.Env.VM.AdminSshPublicKey)

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
	sshPublicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDbFXrLEF2PYNodfwNfGe+4qM3FeZ/FxcfYLZwxStKVW/eTgYn3Y0DQSti86mA+Jrzkx2aSvHDPPJEQUUTiZUMwTiJlR4ud3FBDYZXQVsNhfJO5zduqLpEEdjtFMP/Y3jGpoh+Eq8U08yWRfs1sKay/THS5MoKIprFVU+yIgHsxNcrU2hymTnt+A43dxKHXd4aZXhfW5qHwBZzNBggRXPFb6RpABx2qk9dQGGHFrGp5p0cC2sekXNFg7lV8PEx8pspu+Kf5mSBd1aphRde8ATR61zEDbAGKi1wbGHhrrZ/dAigcSB5YNgllTg5l09CwtjWBFHGY1oxwb+F3foXH19dDIlkB7wsyoT/XD7NMOfNyqDYLlOrVVMPfEdNkBXdCScK8Q3rrT/LL/7fefo/OirDnCvL3dxEA/9ay0hVEHyef6E++tiO9DU/HBVAY6NYjYQCZZ92rqVPzM94ppBU4XocxzAQ7zL+pFABbZkYtXTH8VaE4A1MTgRXvte1CmzeFPQs= emil@thinkpad"

	setup(t)

	client, err := cs.New(&cs.ClientConf{
		URL:         conf.Env.CS.URL,
		ApiKey:      conf.Env.CS.ApiKey,
		Secret:      conf.Env.CS.Secret,
		IpAddressID: conf.Env.CS.IpAddressID,
		NetworkID:   conf.Env.CS.NetworkID,
		ProjectID:   conf.Env.CS.ProjectID,
		ZoneID:      conf.Env.CS.ZoneID,
	})

	if err != nil {
		t.Fatalf(err.Error())
	}

	public := &models.VmPublic{
		Name:              "acc-test-" + uuid.New().String(),
		ServiceOfferingID: "8da28b4d-5fec-4a44-aee7-fb0c5c8265a9", // Small HA
		TemplateID:        "fb6b6b11-6196-42d9-a12d-038bdeecb6f6", // Ubuntu Server
	}

	id, err := client.CreateVM(public, sshPublicKey, conf.Env.VM.AdminSshPublicKey)

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

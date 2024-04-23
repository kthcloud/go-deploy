package cs

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/cs"
	"go-deploy/pkg/subsystems/cs/models"
	"go-deploy/test"
	"go-deploy/test/acc"
	"math/rand"
	"testing"
)

func withClient(t *testing.T) *cs.Client {
	zoneName := "se-flem-2"
	zone := config.Config.VM.GetLegacyZone(zoneName)
	if zone == nil {
		t.Fatalf("no zone with name %s found", zoneName)
	}

	client, err := cs.New(&cs.ClientConf{
		URL:         config.Config.CS.URL,
		ApiKey:      config.Config.CS.ApiKey,
		Secret:      config.Config.CS.Secret,
		ZoneID:      zone.ZoneID,
		ProjectID:   zone.ProjectID,
		IpAddressID: zone.IpAddressID,
		NetworkID:   zone.NetworkID,
		TemplateID:  zone.TemplateID,
	})
	test.NoError(t, err, "failed to create cs client")
	assert.NotNil(t, client, "cs client is nil")

	return client
}

func withDefaultVM(t *testing.T) *models.VmPublic {
	name := acc.GenName()

	defaultVM := &models.VmPublic{
		Name:        name,
		CpuCores:    2,
		RAM:         4,
		ExtraConfig: "",
		Tags: []models.Tag{
			{Key: "name", Value: name},
			{Key: "managedBy", Value: "test"},
			{Key: "deployName", Value: name},
		},
	}

	return withVM(t, defaultVM)
}

func withVM(t *testing.T, vm *models.VmPublic) *models.VmPublic {
	sshPublicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDbFXrLEF2PYNodfwNfGe+4qM3FeZ/FxcfYLZwxStKVW/eTgYn3Y0DQSti86mA+Jrzkx2aSvHDPPJEQUUTiZUMwTiJlR4ud3FBDYZXQVsNhfJO5zduqLpEEdjtFMP/Y3jGpoh+Eq8U08yWRfs1sKay/THS5MoKIprFVU+yIgHsxNcrU2hymTnt+A43dxKHXd4aZXhfW5qHwBZzNBggRXPFb6RpABx2qk9dQGGHFrGp5p0cC2sekXNFg7lV8PEx8pspu+Kf5mSBd1aphRde8ATR61zEDbAGKi1wbGHhrrZ/dAigcSB5YNgllTg5l09CwtjWBFHGY1oxwb+F3foXH19dDIlkB7wsyoT/XD7NMOfNyqDYLlOrVVMPfEdNkBXdCScK8Q3rrT/LL/7fefo/OirDnCvL3dxEA/9ay0hVEHyef6E++tiO9DU/HBVAY6NYjYQCZZ92rqVPzM94ppBU4XocxzAQ7zL+pFABbZkYtXTH8VaE4A1MTgRXvte1CmzeFPQs= emil@thinkpad"

	client := withClient(t)

	client.WithUserSshPublicKey(sshPublicKey)
	client.WithAdminSshPublicKey(config.Config.VM.AdminSshPublicKey)

	vmCreated, err := client.CreateVM(vm)
	test.NoError(t, err, "failed to create vm")
	assert.NotEmpty(t, vmCreated.ID, "no vm id received from client")
	t.Cleanup(func() { cleanUpVM(t, vmCreated.ID) })

	assert.Equal(t, vm.Name, vmCreated.Name, "vm name is not equal")
	assert.Equal(t, vm.CpuCores, vmCreated.CpuCores, "vm cpu cores is not equal")
	assert.Equal(t, vm.RAM, vmCreated.RAM, "vm ram is not equal")
	assert.Equal(t, vm.ExtraConfig, vmCreated.ExtraConfig, "vm extra config is not equal")

	return vmCreated
}

func withDefaultPFR(t *testing.T, vm *models.VmPublic) *models.PortForwardingRulePublic {
	zone := config.Config.VM.GetLegacyZone("se-flem")

	pfr := &models.PortForwardingRulePublic{
		Name:        acc.GenName(),
		VmID:        vm.ID,
		NetworkID:   zone.NetworkID,
		IpAddressID: zone.IpAddressID,
		// Make sure this port is not in the range used by go-deploy
		PublicPort:  10000 + rand.Intn(55000),
		PrivatePort: 22,
		Protocol:    "tcp",
		Tags:        []models.Tag{{Key: "name", Value: acc.GenName()}},
	}

	return withPortForwardingRule(t, pfr)
}

func withPortForwardingRule(t *testing.T, pfr *models.PortForwardingRulePublic) *models.PortForwardingRulePublic {
	client := withClient(t)

	pfrCreated, err := client.CreatePortForwardingRule(pfr)
	test.NoError(t, err, "failed to create port forwarding rule")

	assert.NotEmpty(t, pfrCreated.ID, "no port forwarding rule id received from client")
	t.Cleanup(func() { cleanUpPortForwardingRule(t, pfrCreated.ID) })

	assert.Equal(t, pfr.VmID, pfrCreated.VmID, "port forwarding rule vm id is not equal")
	assert.Equal(t, pfr.NetworkID, pfrCreated.NetworkID, "port forwarding rule network id is not equal")
	assert.Equal(t, pfr.IpAddressID, pfrCreated.IpAddressID, "port forwarding rule ip address id is not equal")
	assert.Equal(t, pfr.PublicPort, pfrCreated.PublicPort, "port forwarding rule public port is not equal")
	assert.Equal(t, pfr.PrivatePort, pfrCreated.PrivatePort, "port forwarding rule private port is not equal")
	assert.Equal(t, pfr.Protocol, pfrCreated.Protocol, "port forwarding rule protocol is not equal")

	return pfrCreated
}

func withDefaultSnapshot(t *testing.T, vm *models.VmPublic) *models.SnapshotPublic {
	snapshot := &models.SnapshotPublic{
		VmID:        vm.ID,
		Name:        acc.GenName(),
		Description: acc.GenName(),
	}

	return withSnapshot(t, snapshot)
}

func withSnapshot(t *testing.T, snapshot *models.SnapshotPublic) *models.SnapshotPublic {
	client := withClient(t)

	snapshotCreated, err := client.CreateSnapshot(snapshot)
	test.NoError(t, err, "failed to create snapshot")
	assert.NotEmpty(t, snapshotCreated.ID, "no snapshot id received from client")
	t.Cleanup(func() { cleanUpSnapshot(t, snapshotCreated.ID) })

	assert.Equal(t, snapshot.VmID, snapshotCreated.VmID, "snapshot vm id is not equal")
	assert.Equal(t, snapshot.Name, snapshotCreated.Name, "snapshot name is not equal")
	assert.Equal(t, snapshot.Description, snapshotCreated.Description, "snapshot description is not equal")

	return snapshotCreated
}

func cleanUpVM(t *testing.T, id string) {
	client := withClient(t)

	err := client.DeleteVM(id)
	test.NoError(t, err, "failed to delete vm")

	deletedVM, err := client.ReadVM(id)
	test.NoError(t, err, "failed to read vm")
	assert.Nil(t, deletedVM, "vm is not nil")

	err = client.DeleteVM(id)
	test.NoError(t, err, "failed to delete vm")
}

func cleanUpPortForwardingRule(t *testing.T, id string) {
	client := withClient(t)

	err := client.DeletePortForwardingRule(id)
	test.NoError(t, err, "failed to delete port forwarding rule")

	deletedPfr, err := client.ReadPortForwardingRule(id)
	test.NoError(t, err, "failed to read port forwarding rule")
	assert.Nil(t, deletedPfr, "port forwarding rule is not nil")

	err = client.DeletePortForwardingRule(id)
	test.NoError(t, err, "failed to delete port forwarding rule")
}

func cleanUpSnapshot(t *testing.T, id string) {
	client := withClient(t)

	err := client.DeleteSnapshot(id)
	test.NoError(t, err, "failed to delete snapshot")

	deletedSnapshot, err := client.ReadSnapshot(id)
	test.NoError(t, err, "failed to read snapshot")
	assert.Nil(t, deletedSnapshot, "snapshot is not nil")

	err = client.DeleteSnapshot(id)
	test.NoError(t, err, "failed to delete snapshot")
}

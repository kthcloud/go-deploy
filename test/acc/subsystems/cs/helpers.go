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
	zoneName := "se-flem"
	zone := config.Config.VM.GetZone(zoneName)
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
	})
	test.NoError(t, err, "failed to create cs client")
	assert.NotNil(t, client, "cs client is nil")

	return client
}

// 2 CPU cores, 1 GB RAM, 25 GB disk
func withCsServiceOfferingSmall(t *testing.T) *models.ServiceOfferingPublic {
	so := &models.ServiceOfferingPublic{
		Name:        acc.GenName(),
		Description: acc.GenName(),
		CpuCores:    2,
		RAM:         1,
		DiskSize:    25,
	}

	return withServiceOffering(t, so)
}

// 4 CPU cores, 2 GB RAM, 25 GB disk
func withCsServiceOfferingBig(t *testing.T) *models.ServiceOfferingPublic {
	so := &models.ServiceOfferingPublic{
		Name:        acc.GenName(),
		Description: acc.GenName(),
		CpuCores:    4,
		RAM:         2,
		DiskSize:    25,
	}

	return withServiceOffering(t, so)
}

func withServiceOffering(t *testing.T, so *models.ServiceOfferingPublic) *models.ServiceOfferingPublic {
	client := withClient(t)

	soCreated, err := client.CreateServiceOffering(so)
	test.NoError(t, err, "failed to create service offering")
	assert.NotEmpty(t, soCreated.ID, "no service offering id received from client")
	t.Cleanup(func() { cleanUpServiceOffering(t, soCreated.ID) })

	assert.Equal(t, so.Name, soCreated.Name, "service offering name is not equal")
	assert.Equal(t, so.Description, soCreated.Description, "service offering description is not equal")
	assert.Equal(t, so.CpuCores, soCreated.CpuCores, "service offering cpu cores is not equal")
	assert.Equal(t, so.RAM, soCreated.RAM, "service offering ram is not equal")
	assert.Equal(t, so.DiskSize, soCreated.DiskSize, "service offering disk size is not equal")

	return soCreated
}

func withDefaultVM(t *testing.T, so *models.ServiceOfferingPublic) *models.VmPublic {
	name := acc.GenName()

	defaultVM := &models.VmPublic{
		Name:              name,
		ServiceOfferingID: so.ID,
		TemplateID:        "cbac58b6-336b-49ab-b4d7-341586dfefcc", // ubuntu-2204-cloudstack-ready-v1.2
		ExtraConfig:       "",
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
	assert.Equal(t, vm.ServiceOfferingID, vmCreated.ServiceOfferingID, "vm service offering id is not equal")
	assert.Equal(t, vm.TemplateID, vmCreated.TemplateID, "vm template id is not equal")
	assert.Equal(t, vm.ExtraConfig, vmCreated.ExtraConfig, "vm extra config is not equal")

	return vmCreated
}

func withDefaultPFR(t *testing.T, vm *models.VmPublic) *models.PortForwardingRulePublic {
	zone := config.Config.VM.GetZone("se-flem")

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

func cleanUpServiceOffering(t *testing.T, id string) {
	client := withClient(t)

	err := client.DeleteServiceOffering(id)
	test.NoError(t, err, "failed to delete service offering")

	deletedServiceOffering, err := client.ReadServiceOffering(id)
	test.NoError(t, err, "failed to read service offering")
	assert.Nil(t, deletedServiceOffering, "service offering is not nil")

	err = client.DeleteServiceOffering(id)
	test.NoError(t, err, "failed to delete service offering")
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

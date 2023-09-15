package cs

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/cs"
	"go-deploy/pkg/subsystems/cs/models"
	"testing"
)

func withCsClient(t *testing.T) *cs.Client {
	zoneName := "se-flem"
	zone := conf.Env.VM.GetZone(zoneName)
	if zone == nil {
		t.Fatalf("no zone with name %s found", zoneName)
	}

	client, err := cs.New(&cs.ClientConf{
		URL:         conf.Env.CS.URL,
		ApiKey:      conf.Env.CS.ApiKey,
		Secret:      conf.Env.CS.Secret,
		ZoneID:      zone.ZoneID,
		ProjectID:   zone.ProjectID,
		IpAddressID: zone.IpAddressID,
		NetworkID:   zone.NetworkID,
	})
	assert.NoError(t, err, "failed to create cs client")
	assert.NotNil(t, client, "cs client is nil")

	return client
}

func withCsServiceOfferingType1(t *testing.T) *models.ServiceOfferingPublic {
	client := withCsClient(t)

	id, err := client.CreateServiceOffering(&models.ServiceOfferingPublic{
		Name:        "acc-test-" + uuid.New().String(),
		Description: "acc-test-" + uuid.New().String(),
		CpuCores:    2,
		RAM:         1,
		DiskSize:    25,
	})

	assert.NoError(t, err, "failed to create service offering")
	assert.NotZero(t, id, "no service offering id received from client")

	so, err := client.ReadServiceOffering(id)
	assert.NoError(t, err, "failed to read service offering")
	assert.NotNil(t, so, "service offering is nil")

	return so
}

func withServiceOfferingType2(t *testing.T) *models.ServiceOfferingPublic {
	client := withCsClient(t)

	id, err := client.CreateServiceOffering(&models.ServiceOfferingPublic{
		Name:        "acc-test-" + uuid.New().String(),
		Description: "acc-test-" + uuid.New().String(),
		CpuCores:    4,
		RAM:         2,
		DiskSize:    25,
	})

	assert.NoError(t, err, "failed to create service offering")
	assert.NotZero(t, id, "no service offering id received from client")

	so, err := client.ReadServiceOffering(id)
	assert.NoError(t, err, "failed to read service offering")
	assert.NotNil(t, so, "service offering is nil")

	return so
}

func withVM(t *testing.T, so *models.ServiceOfferingPublic) *models.VmPublic {
	sshPublicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDbFXrLEF2PYNodfwNfGe+4qM3FeZ/FxcfYLZwxStKVW/eTgYn3Y0DQSti86mA+Jrzkx2aSvHDPPJEQUUTiZUMwTiJlR4ud3FBDYZXQVsNhfJO5zduqLpEEdjtFMP/Y3jGpoh+Eq8U08yWRfs1sKay/THS5MoKIprFVU+yIgHsxNcrU2hymTnt+A43dxKHXd4aZXhfW5qHwBZzNBggRXPFb6RpABx2qk9dQGGHFrGp5p0cC2sekXNFg7lV8PEx8pspu+Kf5mSBd1aphRde8ATR61zEDbAGKi1wbGHhrrZ/dAigcSB5YNgllTg5l09CwtjWBFHGY1oxwb+F3foXH19dDIlkB7wsyoT/XD7NMOfNyqDYLlOrVVMPfEdNkBXdCScK8Q3rrT/LL/7fefo/OirDnCvL3dxEA/9ay0hVEHyef6E++tiO9DU/HBVAY6NYjYQCZZ92rqVPzM94ppBU4XocxzAQ7zL+pFABbZkYtXTH8VaE4A1MTgRXvte1CmzeFPQs= emil@thinkpad"

	client := withCsClient(t)

	name := "acc-test-" + uuid.New().String()
	vm := &models.VmPublic{
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

	id, err := client.CreateVM(vm, sshPublicKey, conf.Env.VM.AdminSshPublicKey)
	assert.NoError(t, err, "failed to create vm")
	assert.NotZero(t, id, "no vm id received from client")

	createdVM, err := client.ReadVM(id)
	assert.NoError(t, err, "failed to read vm")
	assert.NotNil(t, vm, "vm is nil")

	vm.ID = id
	assert.NotEmpty(t, vm.ID, "vm id is empty")
	assert.Equal(t, vm.Name, createdVM.Name, "vm name is not equal")

	assert.EqualValues(t, vm.ServiceOfferingID, createdVM.ServiceOfferingID, "vm service offering id is not equal")

	return vm
}

func withPortForwardingRule(t *testing.T, vm *models.VmPublic) *models.PortForwardingRulePublic {
	client := withCsClient(t)

	pfr := &models.PortForwardingRulePublic{
		VmID:        vm.ID,
		Protocol:    "tcp",
		PrivatePort: 22,
		PublicPort:  2222,
	}

	id, err := client.CreatePortForwardingRule(pfr)
	assert.NoError(t, err, "failed to create port forwarding rule")
	assert.NotZero(t, id, "no port forwarding rule id received from client")

	createdPfr, err := client.ReadPortForwardingRule(id)
	assert.NoError(t, err, "failed to read port forwarding rule")
	assert.NotNil(t, pfr, "port forwarding rule is nil")

	pfr.ID = id
	assert.NotEmpty(t, pfr.ID, "port forwarding rule id is zero")
	assert.Equal(t, pfr, createdPfr, "port forwarding rule is not equal")

	assert.EqualValues(t, pfr, createdPfr, "port forwarding rule is not equal")

	return pfr
}

func cleanUpServiceOffering(t *testing.T, id string) {
	client := withCsClient(t)

	err := client.DeleteServiceOffering(id)
	assert.NoError(t, err, "failed to delete service offering")

	deletedServiceOffering, err := client.ReadServiceOffering(id)
	assert.NoError(t, err, "failed to read service offering")
	assert.Nil(t, deletedServiceOffering, "service offering is not nil")

	err = client.DeleteServiceOffering(id)
	assert.NoError(t, err, "failed to delete service offering")
}

func cleanUpVM(t *testing.T, id string) {
	client := withCsClient(t)

	err := client.DeleteVM(id)
	assert.NoError(t, err, "failed to delete vm")

	deletedVM, err := client.ReadVM(id)
	assert.NoError(t, err, "failed to read vm")
	assert.Nil(t, deletedVM, "vm is not nil")

	err = client.DeleteVM(id)
	assert.NoError(t, err, "failed to delete vm")
}

func cleanUpPortForwardingRule(t *testing.T, id string) {
	client := withCsClient(t)

	err := client.DeletePortForwardingRule(id)
	assert.NoError(t, err, "failed to delete port forwarding rule")

	deletedPfr, err := client.ReadPortForwardingRule(id)
	assert.NoError(t, err, "failed to read port forwarding rule")
	assert.Nil(t, deletedPfr, "port forwarding rule is not nil")

	err = client.DeletePortForwardingRule(id)
	assert.NoError(t, err, "failed to delete port forwarding rule")
}

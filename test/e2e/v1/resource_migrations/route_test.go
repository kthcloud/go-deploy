package resource_migrations

import (
	"github.com/stretchr/testify/assert"
	bodyV2 "go-deploy/dto/v2/body"
	"go-deploy/models/model"
	"go-deploy/test/e2e"
	v1 "go-deploy/test/e2e/v1"
	v2 "go-deploy/test/e2e/v2"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	e2e.Setup()
	code := m.Run()
	e2e.Shutdown()
	os.Exit(code)
}

func TestList(t *testing.T) {
	t.Parallel()

	queries := []string{
		"?page=1&pageSize=10",
		"?userId=" + model.TestPowerUserID + "&page=1&pageSize=3",
	}

	for _, query := range queries {
		v1.ListResourceMigrations(t, query)
	}
}

func TestUpdateDeploymentOwnerNonAdmin(t *testing.T) {
	t.Parallel()

	resource1, _ := v1.WithDeployment(t, bodyV2.DeploymentCreate{
		Name:    e2e.GenName(),
		Private: false,
	}, e2e.PowerUser)

	// Create as default user
	resourceMigration1 := v1.WithResourceMigration(t, bodyV2.ResourceMigrationCreate{
		Type:       model.ResourceMigrationTypeUpdateOwner,
		ResourceID: resource1.ID,
		UpdateOwner: &struct {
			OwnerID string `json:"ownerId" binding:"required,uuid4"`
		}{
			OwnerID: model.TestDefaultUserID,
		},
	}, e2e.PowerUser)

	// They generate notifications for the receiver, find them
	notifications := v1.ListNotifications(t, "", e2e.DefaultUser)
	var notification1 *bodyV2.NotificationRead
	//var notification2 *body.NotificationRead
	for _, n := range notifications {
		no := n
		if n.Content["id"] == resourceMigration1.ResourceID {
			notification1 = &no
		}
	}

	e2e.MustNotNil(t, notification1, "notification1 not found")

	// Get the codes
	code1, ok := notification1.Content["code"].(string)
	assert.True(t, ok, "code1 not found")

	// Accept the migrations by updating their status to accepted
	v1.UpdateResourceMigration(t, resourceMigration1.ID, bodyV2.ResourceMigrationUpdate{
		Status: model.ResourceMigrationStatusAccepted,
		Code:   &code1,
	}, e2e.DefaultUser)

	// Check if the owner was updated
	resource1 = v1.GetDeployment(t, resource1.ID, e2e.DefaultUser)
	assert.Equal(t, model.TestDefaultUserID, resource1.OwnerID, "deployment owner not updated")

	// Ensure resources are running
	v1.WaitForDeploymentRunning(t, resource1.ID, func(d *bodyV2.DeploymentRead) bool {
		if d.URL != nil {
			return v1.CheckUpURL(t, *d.URL)
		}
		return false
	})
}

func TestUpdateVmOwnerNonAdmin(t *testing.T) {
	t.Parallel()

	resource2 := v2.WithVM(t, bodyV2.VmCreate{
		Name:         e2e.GenName(),
		SshPublicKey: v2.WithSshPublicKey(t),
		CpuCores:     1,
		RAM:          4,
		DiskSize:     10,
	}, e2e.PowerUser)

	// Create as default user
	resourceMigration2 := v1.WithResourceMigration(t, bodyV2.ResourceMigrationCreate{
		Type:       model.ResourceMigrationTypeUpdateOwner,
		ResourceID: resource2.ID,
		UpdateOwner: &struct {
			OwnerID string `json:"ownerId" binding:"required,uuid4"`
		}{
			OwnerID: model.TestDefaultUserID,
		},
	}, e2e.PowerUser)

	// They generate notifications for the receiver, find them
	notifications := v1.ListNotifications(t, "", e2e.DefaultUser)
	var notification2 *bodyV2.NotificationRead
	for _, n := range notifications {
		no := n
		if n.Content["id"] == resourceMigration2.ResourceID {
			notification2 = &no
		}
	}

	e2e.MustNotNil(t, notification2, "notification2 not found")

	// Get the codes
	code2, ok := notification2.Content["code"].(string)
	assert.True(t, ok, "code2 not found")

	// Accept the migrations by updating their status to accepted
	v1.UpdateResourceMigration(t, resourceMigration2.ID, bodyV2.ResourceMigrationUpdate{
		Status: model.ResourceMigrationStatusAccepted,
		Code:   &code2,
	}, e2e.DefaultUser)

	// Check if the owner was updated
	resource2 = v2.GetVM(t, resource2.ID, e2e.DefaultUser)
	assert.Equal(t, model.TestDefaultUserID, resource2.OwnerID, "vm owner not updated")

	// Ensure resources are running
	v2.WaitForVmRunning(t, resource2.ID, func(vm *bodyV2.VmRead) bool {
		if vm.SshConnectionString != nil {
			res := v2.DoSshCommand(t, *vm.SshConnectionString, "echo 'hello world'")
			if strings.Contains(res, "hello world") {
				return true
			}
		}
		return false
	})
}

func TestUpdateDeploymentOwnerAsAdmin(t *testing.T) {
	t.Parallel()

	resource1, _ := v1.WithDeployment(t, bodyV2.DeploymentCreate{
		Name:    e2e.GenName(),
		Private: false,
	}, e2e.PowerUser)

	// Create as default user
	s := model.ResourceMigrationStatusAccepted
	v1.WithResourceMigration(t, bodyV2.ResourceMigrationCreate{
		Type:       model.ResourceMigrationTypeUpdateOwner,
		ResourceID: resource1.ID,
		Status:     &s,
		UpdateOwner: &struct {
			OwnerID string `json:"ownerId" binding:"required,uuid4"`
		}{
			OwnerID: model.TestDefaultUserID,
		},
	}, e2e.AdminUser)

	// They should be instant now
	resource1 = v1.GetDeployment(t, resource1.ID, e2e.DefaultUser)
	assert.Equal(t, model.TestDefaultUserID, resource1.OwnerID, "deployment owner not updated")

	// Ensure resource are accessible by the new owner
	v1.WaitForDeploymentRunning(t, resource1.ID, func(d *bodyV2.DeploymentRead) bool {
		if d.URL != nil {
			return v1.CheckUpURL(t, *d.URL)
		}
		return false
	})
}

func TestUpdateVmOwnerAsAdmin(t *testing.T) {
	t.Parallel()

	resource2 := v2.WithVM(t, bodyV2.VmCreate{
		Name:         e2e.GenName(),
		SshPublicKey: v2.WithSshPublicKey(t),
		CpuCores:     1,
		RAM:          4,
		DiskSize:     10,
	}, e2e.PowerUser)

	// Create as default user
	s := model.ResourceMigrationStatusAccepted
	v1.WithResourceMigration(t, bodyV2.ResourceMigrationCreate{
		Type:       model.ResourceMigrationTypeUpdateOwner,
		ResourceID: resource2.ID,
		Status:     &s,
		UpdateOwner: &struct {
			OwnerID string `json:"ownerId" binding:"required,uuid4"`
		}{
			OwnerID: model.TestDefaultUserID,
		},
	}, e2e.AdminUser)

	// They should be instant now
	resource2 = v2.GetVM(t, resource2.ID, e2e.DefaultUser)
	assert.Equal(t, model.TestDefaultUserID, resource2.OwnerID, "vm owner not updated")

	// Ensure resource are accessible by the new owner
	v2.WaitForVmRunning(t, resource2.ID, func(vm *bodyV2.VmRead) bool {
		if vm.SshConnectionString != nil {
			res := v2.DoSshCommand(t, *vm.SshConnectionString, "echo 'hello world'")
			if strings.Contains(res, "hello world") {
				return true
			}
		}
		return false
	})
}

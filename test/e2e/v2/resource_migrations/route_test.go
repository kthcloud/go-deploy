package resource_migrations

import (
	"github.com/stretchr/testify/assert"
	bodyV2 "github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/test/e2e"
	v2 "github.com/kthcloud/go-deploy/test/e2e/v2"
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
		v2.ListResourceMigrations(t, query)
	}
}

func TestUpdateDeploymentOwnerNonAdmin(t *testing.T) {
	t.Parallel()

	resource, _ := v2.WithDeployment(t, bodyV2.DeploymentCreate{
		Name:    e2e.GenName(),
		Private: false,
	}, e2e.PowerUser)

	// Create as default user
	_ = v2.WithResourceMigration(t, bodyV2.ResourceMigrationCreate{
		Type:       model.ResourceMigrationTypeUpdateOwner,
		ResourceID: resource.ID,
		UpdateOwner: &struct {
			OwnerID string `json:"ownerId" binding:"required,uuid4"`
		}{
			OwnerID: model.TestDefaultUserID,
		},
	}, e2e.PowerUser)

	// They generate notifications for the receiver, find them
	notifications := v2.ListNotifications(t, "", e2e.DefaultUser)
	var notification *bodyV2.NotificationRead
	for _, n := range notifications {
		no := n
		if n.Content["resourceId"] == resource.ID && n.Type == model.NotificationResourceTransfer {
			notification = &no
		}
	}

	e2e.MustNotNil(t, notification, "notification not found")

	// Get the codes
	code, ok := notification.Content["code"].(string)
	assert.True(t, ok, "code not found")

	// Accept the migrations by updating their status to accepted
	v2.UpdateResourceMigration(t, notification.Content["id"].(string), bodyV2.ResourceMigrationUpdate{
		Status: model.ResourceMigrationStatusAccepted,
		Code:   &code,
	}, e2e.DefaultUser)

	// Check if the owner was updated
	resource = v2.GetDeployment(t, resource.ID, e2e.DefaultUser)
	assert.Equal(t, model.TestDefaultUserID, resource.OwnerID, "deployment owner not updated")

	// Ensure resources are running
	v2.WaitForDeploymentRunning(t, resource.ID, func(d *bodyV2.DeploymentRead) bool {
		if d.URL != nil {
			return v2.CheckUpURL(t, *d.URL)
		}
		return false
	})
}

func TestUpdateVmOwnerNonAdmin(t *testing.T) {
	t.Parallel()

	if !e2e.VmTestsEnabled {
		t.Skip("vm tests are disabled")
	}

	resource := v2.WithVM(t, bodyV2.VmCreate{
		Name:         e2e.GenName(),
		SshPublicKey: v2.WithSshPublicKey(t),
		CpuCores:     1,
		RAM:          4,
		DiskSize:     10,
	}, e2e.PowerUser)

	// Create as default user
	_ = v2.WithResourceMigration(t, bodyV2.ResourceMigrationCreate{
		Type:       model.ResourceMigrationTypeUpdateOwner,
		ResourceID: resource.ID,
		UpdateOwner: &struct {
			OwnerID string `json:"ownerId" binding:"required,uuid4"`
		}{
			OwnerID: model.TestDefaultUserID,
		},
	}, e2e.PowerUser)

	// They generate notifications for the receiver, find them
	notifications := v2.ListNotifications(t, "", e2e.DefaultUser)
	var notification *bodyV2.NotificationRead
	for _, n := range notifications {
		no := n
		if n.Content["resourceId"] == resource.ID && n.Type == model.NotificationResourceTransfer {
			notification = &no
		}
	}

	e2e.MustNotNil(t, notification, "notification not found")

	// Get the codes
	code, ok := notification.Content["code"].(string)
	assert.True(t, ok, "code not found")

	// Accept the migrations by updating their status to accepted
	v2.UpdateResourceMigration(t, notification.Content["id"].(string), bodyV2.ResourceMigrationUpdate{
		Status: model.ResourceMigrationStatusAccepted,
		Code:   &code,
	}, e2e.DefaultUser)

	// Check if the owner was updated
	resource = v2.GetVM(t, resource.ID, e2e.DefaultUser)
	assert.Equal(t, model.TestDefaultUserID, resource.OwnerID, "vm owner not updated")

	// Ensure resources are running
	v2.WaitForVmRunning(t, resource.ID, func(vm *bodyV2.VmRead) bool {
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

	resource1, _ := v2.WithDeployment(t, bodyV2.DeploymentCreate{
		Name:    e2e.GenName(),
		Private: false,
	}, e2e.PowerUser)

	// Create as default user
	s := model.ResourceMigrationStatusAccepted
	v2.WithResourceMigration(t, bodyV2.ResourceMigrationCreate{
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
	resource1 = v2.GetDeployment(t, resource1.ID, e2e.DefaultUser)
	assert.Equal(t, model.TestDefaultUserID, resource1.OwnerID, "deployment owner not updated")

	// Ensure resource are accessible by the new owner
	v2.WaitForDeploymentRunning(t, resource1.ID, func(d *bodyV2.DeploymentRead) bool {
		if d.URL != nil {
			return v2.CheckUpURL(t, *d.URL)
		}
		return false
	})
}

func TestUpdateVmOwnerAsAdmin(t *testing.T) {
	t.Parallel()

	if !e2e.VmTestsEnabled {
		t.Skip("vm tests are disabled")
	}

	resource2 := v2.WithVM(t, bodyV2.VmCreate{
		Name:         e2e.GenName(),
		SshPublicKey: v2.WithSshPublicKey(t),
		CpuCores:     1,
		RAM:          4,
		DiskSize:     10,
	}, e2e.PowerUser)

	// Create as default user
	s := model.ResourceMigrationStatusAccepted
	v2.WithResourceMigration(t, bodyV2.ResourceMigrationCreate{
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

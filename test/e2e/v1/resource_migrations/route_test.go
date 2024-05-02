package resource_migrations

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/dto/v1/body"
	bodyV2 "go-deploy/dto/v2/body"
	"go-deploy/models/model"
	"go-deploy/test/e2e"
	v1 "go-deploy/test/e2e/v1"
	v2 "go-deploy/test/e2e/v2"
	"os"
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
		"?userId=" + e2e.PowerUserID + "&page=1&pageSize=3",
	}

	for _, query := range queries {
		v1.ListResourceMigrations(t, query)
	}
}

func TestCreateOwnerUpdateNoAdmin(t *testing.T) {
	t.Parallel()

	resource1, _ := v1.WithDeployment(t, body.DeploymentCreate{
		Name:    e2e.GenName(),
		Private: false,
	}, e2e.DefaultUserID)

	resource2 := v2.WithVM(t, bodyV2.VmCreate{
		Name:         e2e.GenName(),
		SshPublicKey: v2.WithSshPublicKey(t),
		CpuCores:     1,
		RAM:          4,
		DiskSize:     10,
	}, e2e.DefaultUserID)

	// Create as default user
	resourceMigration1 := v1.WithResourceMigration(t, body.ResourceMigrationCreate{
		Type:       model.ResourceMigrationTypeUpdateOwner,
		ResourceID: resource1.ID,
		UpdateOwner: &struct {
			OwnerID string `json:"ownerId" binding:"required,uuid4"`
		}{
			OwnerID: e2e.PowerUserID,
		},
	}, e2e.DefaultUserID)
	resourceMigration2 := v1.WithResourceMigration(t, body.ResourceMigrationCreate{
		Type:       model.ResourceMigrationTypeUpdateOwner,
		ResourceID: resource2.ID,
		UpdateOwner: &struct {
			OwnerID string `json:"ownerId" binding:"required,uuid4"`
		}{
			OwnerID: e2e.PowerUserID,
		},
	}, e2e.DefaultUserID)

	// They generate notifications for the receiver, find them
	notifications := v1.ListNotifications(t, "", e2e.PowerUserID)
	var notification1 *body.NotificationRead
	var notification2 *body.NotificationRead
	for _, n := range notifications {
		no := n
		if n.Content["id"] == resourceMigration1.ResourceID {
			notification1 = &no
		}

		if n.Content["id"] == resourceMigration2.ResourceID {
			notification2 = &no
		}
	}
	assert.NotNil(t, notification1, "notification1 not found")
	assert.NotNil(t, notification2, "notification2 not found")

	// Get the codes
	code1, ok := notification1.Content["code"].(string)
	assert.True(t, ok, "code1 not found")
	code2, ok := notification2.Content["code"].(string)
	assert.True(t, ok, "code2 not found")

	// Accept the migrations by updating their status to accepted
	v1.UpdateResourceMigration(t, resourceMigration1.ID, body.ResourceMigrationUpdate{
		Status: model.ResourceMigrationStatusAccepted,
		Code:   &code1,
	}, e2e.PowerUserID)
	v1.UpdateResourceMigration(t, resourceMigration2.ID, body.ResourceMigrationUpdate{
		Status: model.ResourceMigrationStatusAccepted,
		Code:   &code2,
	}, e2e.PowerUserID)

	// Check if the owner was updated
	resource1 = v1.GetDeployment(t, resource1.ID)
	assert.Equal(t, e2e.PowerUserID, resource1.OwnerID, "deployment owner not updated")
	resource2 = v2.GetVM(t, resource2.ID)
	assert.Equal(t, e2e.PowerUserID, resource2.OwnerID, "vm owner not updated")
}

func TestCreateOwnerUpdateAsAdmin(t *testing.T) {
	t.Parallel()

	resource1, _ := v1.WithDeployment(t, body.DeploymentCreate{
		Name:    e2e.GenName(),
		Private: false,
	}, e2e.AdminUserID)

	resource2 := v2.WithVM(t, bodyV2.VmCreate{
		Name:         e2e.GenName(),
		SshPublicKey: v2.WithSshPublicKey(t),
		CpuCores:     1,
		RAM:          4,
		DiskSize:     10,
	}, e2e.AdminUserID)

	// Create as default user
	s := model.ResourceMigrationStatusAccepted
	v1.WithResourceMigration(t, body.ResourceMigrationCreate{
		Type:       model.ResourceMigrationTypeUpdateOwner,
		ResourceID: resource1.ID,
		Status:     &s,
		UpdateOwner: &struct {
			OwnerID string `json:"ownerId" binding:"required,uuid4"`
		}{
			OwnerID: e2e.PowerUserID,
		},
	}, e2e.AdminUserID)
	v1.WithResourceMigration(t, body.ResourceMigrationCreate{
		Type:       model.ResourceMigrationTypeUpdateOwner,
		ResourceID: resource2.ID,
		Status:     &s,
		UpdateOwner: &struct {
			OwnerID string `json:"ownerId" binding:"required,uuid4"`
		}{
			OwnerID: e2e.PowerUserID,
		},
	}, e2e.AdminUserID)

	// They should be instant now
	resource1 = v1.GetDeployment(t, resource1.ID)
	assert.Equal(t, e2e.PowerUserID, resource1.OwnerID, "deployment owner not updated")
	resource2 = v2.GetVM(t, resource2.ID)
	assert.Equal(t, e2e.PowerUserID, resource2.OwnerID, "vm owner not updated")
}

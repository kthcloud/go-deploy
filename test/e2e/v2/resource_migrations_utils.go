//go:build e2e

package v2

import (
	body "github.com/kthcloud/go-deploy/api/openapi/v2"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/test/e2e"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

const (
	ResourceMigrationPath  = "/v2/resourceMigrations/"
	ResourceMigrationsPath = "/v2/resourceMigrations"
)

func GetResourceMigration(t *testing.T, id string, user ...string) body.ResourceMigrationRead {
	resp := e2e.DoGetRequest(t, ResourceMigrationPath+id, user...)
	return e2e.MustParse[body.ResourceMigrationRead](t, resp)
}

func ResourceMigrationExists(t *testing.T, id string, user ...string) bool {
	resp := e2e.DoGetRequest(t, ResourceMigrationPath+id, user...)
	return resp.StatusCode == http.StatusOK
}

func ListResourceMigrations(t *testing.T, query string, user ...string) []body.ResourceMigrationRead {
	resp := e2e.DoGetRequest(t, ResourceMigrationsPath+query, user...)
	return e2e.MustParse[[]body.ResourceMigrationRead](t, resp)
}

func UpdateResourceMigration(t *testing.T, id string, requestBody body.ResourceMigrationUpdate, user ...string) body.ResourceMigrationRead {
	resp := e2e.DoPostRequest(t, ResourceMigrationPath+id, requestBody, user...)
	resourceMigrationUpdated := e2e.MustParse[body.ResourceMigrationUpdated](t, resp)

	// It is either done immediately or has a job
	if resourceMigrationUpdated.JobID != nil {
		WaitForJobFinished(t, *resourceMigrationUpdated.JobID, nil)
	}

	assert.Equal(t, requestBody.Status, resourceMigrationUpdated.Status)

	return body.ResourceMigrationRead{
		ID:           resourceMigrationUpdated.ID,
		ResourceID:   resourceMigrationUpdated.ResourceID,
		UserID:       resourceMigrationUpdated.UserID,
		Type:         resourceMigrationUpdated.Type,
		ResourceType: resourceMigrationUpdated.ResourceType,
		Status:       resourceMigrationUpdated.Status,
		UpdateOwner:  resourceMigrationUpdated.UpdateOwner,
		CreatedAt:    resourceMigrationUpdated.CreatedAt,
		DeletedAt:    resourceMigrationUpdated.DeletedAt,
	}
}

func DeleteResourceMigration(t *testing.T, id string, user ...string) {
	e2e.DoDeleteRequest(t, ResourceMigrationPath+id, user...)
}

func WithResourceMigration(t *testing.T, requestBody body.ResourceMigrationCreate, user ...string) body.ResourceMigrationRead {
	resp := e2e.DoPostRequest(t, ResourceMigrationsPath, requestBody, user...)
	resourceMigrationCreated := e2e.MustParse[body.ResourceMigrationCreated](t, resp)
	t.Cleanup(func() {
		CleanUpResourceMigration(t, resourceMigrationCreated.ID)
	})

	if resourceMigrationCreated.JobID != nil {
		WaitForJobFinished(t, *resourceMigrationCreated.JobID, nil)
	}

	assert.NotEmpty(t, resourceMigrationCreated.ID)
	assert.Equal(t, requestBody.ResourceID, resourceMigrationCreated.ResourceID)
	assert.Equal(t, requestBody.Type, resourceMigrationCreated.Type)
	if requestBody.Status != nil {
		assert.Equal(t, *requestBody.Status, resourceMigrationCreated.Status)
	} else {
		assert.Equal(t, model.ResourceMigrationStatusPending, resourceMigrationCreated.Status)
	}

	if requestBody.UpdateOwner != nil {
		assert.Equal(t, requestBody.UpdateOwner.OwnerID, resourceMigrationCreated.UpdateOwner.OwnerID)
	}

	return body.ResourceMigrationRead{
		ID:           resourceMigrationCreated.ID,
		ResourceID:   resourceMigrationCreated.ResourceID,
		UserID:       resourceMigrationCreated.UserID,
		Type:         resourceMigrationCreated.Type,
		ResourceType: resourceMigrationCreated.ResourceType,
		Status:       resourceMigrationCreated.Status,
		UpdateOwner:  resourceMigrationCreated.UpdateOwner,
		CreatedAt:    resourceMigrationCreated.CreatedAt,
		DeletedAt:    resourceMigrationCreated.DeletedAt,
	}
}

func CleanUpResourceMigration(t *testing.T, id string) {
	resp := e2e.DoDeleteRequest(t, ResourceMigrationPath+id, e2e.AdminUser)
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusNoContent {
		return
	}

	t.Fatalf("resource migration was not deleted")
}

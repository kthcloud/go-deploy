package sms

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/dto/v1/body"
	"go-deploy/test/e2e"
	"os"
	"testing"
	"time"
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
	}

	for _, query := range queries {
		e2e.ListSMs(t, query)
	}
}

func TestCreate(t *testing.T) {
	t.Parallel()

	// To test this, we should just list deployments, since this triggers creating it
	e2e.DoGetRequest(t, "/deployments")

	// Make sure the storage manager has time to be created
	time.Sleep(30 * time.Second)

	storageManagers := e2e.ListSMs(t, "?all=false")
	assert.NotEmpty(t, storageManagers, "storage managers were empty")

	storageManager := storageManagers[0]
	assert.NotEmpty(t, storageManager.ID, "storage manager id was empty")
	assert.NotEmpty(t, storageManager.OwnerID, "storage manager owner id was empty")
	assert.NotEmpty(t, storageManager.URL, "storage manager url was empty")

	e2e.WaitForSmRunning(t, storageManager.ID, func(storageManagerRead *body.SmRead) bool {
		// Make sure it is accessible
		if storageManager.URL != nil {
			return e2e.CheckUpURL(t, *storageManager.URL)
		}
		return false
	})

	// Ensure the User has the storage url
	user := e2e.GetUser(t, e2e.AdminUserID)
	assert.NotEmpty(t, user.StorageURL, "storage url was empty")
}

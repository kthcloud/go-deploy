package sms

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/body"
	"go-deploy/test/e2e"
	"net/http"
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
	resp := e2e.DoGetRequest(t, "/storageManagers")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var storageManagers []body.SmRead
	err := e2e.ReadResponseBody(t, resp, &storageManagers)
	assert.NoError(t, err, "storage managers were not fetched")

	for _, storageManager := range storageManagers {
		assert.NotEmpty(t, storageManager.ID, "storage manager id was empty")
		assert.NotEmpty(t, storageManager.OwnerID, "storage manager owner id was empty")
		assert.NotEmpty(t, storageManager.URL, "storage manager url was empty")
	}
}

func TestCreate(t *testing.T) {
	// To test this, we should just list deployments, since this triggers creating it
	_ = e2e.DoGetRequest(t, "/deployments")

	// Make sure the storage manager has time to be created
	time.Sleep(30 * time.Second)

	storageManager := e2e.GetSM(t, e2e.AdminUserID)
	assert.NotEmpty(t, storageManager.ID, "storage manager id was empty")
	assert.NotEmpty(t, storageManager.OwnerID, "storage manager owner id was empty")
	assert.NotEmpty(t, storageManager.URL, "storage manager url was empty")

	e2e.WaitForSmRunning(t, storageManager.ID, func(storageManagerRead *body.SmRead) bool {
		//make sure it is accessible
		if storageManager.URL != nil {
			return e2e.CheckUpURL(t, *storageManager.URL)
		}
		return false
	})

	// Ensure the User has the storage url
	user := e2e.GetUser(t, e2e.AdminUserID)
	assert.NotEmpty(t, user.StorageURL, "storage url was empty")
}

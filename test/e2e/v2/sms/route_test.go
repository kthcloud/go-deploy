package sms

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/dto/v2/body"
	"go-deploy/models/model"
	"go-deploy/test/e2e"
	"go-deploy/test/e2e/v2"
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
		v2.ListSMs(t, query)
	}
}

func TestCreate(t *testing.T) {
	t.Parallel()

	v2.ListDeployments(t, "?all=true")

	// Make sure the storage manager has time to be created
	time.Sleep(5 * time.Second)

	storageManagers := v2.ListSMs(t, "?all=false")
	assert.NotEmpty(t, storageManagers, "storage managers were empty")

	storageManager := storageManagers[0]
	assert.NotEmpty(t, storageManager.ID, "storage manager id was empty")
	assert.NotEmpty(t, storageManager.OwnerID, "storage manager owner id was empty")
	assert.NotEmpty(t, storageManager.URL, "storage manager url was empty")

	v2.WaitForSmRunning(t, storageManager.ID, func(storageManagerRead *body.SmRead) bool {
		// Make sure it is accessible
		if storageManager.URL != nil {
			return v2.CheckUpURL(t, *storageManager.URL)
		}
		return false
	})

	// Ensure the User has the storage url
	user := v2.GetUser(t, model.TestPowerUserID)
	assert.NotEmpty(t, user.StorageURL, "storage url was empty")
}

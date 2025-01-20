package sms

import (
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/test/e2e"
	"github.com/kthcloud/go-deploy/test/e2e/v2"
	"github.com/stretchr/testify/assert"
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

	v2.WithDeployment(t, body.DeploymentCreate{Name: e2e.GenName("sms-create")}, e2e.PowerUser)

	// Make sure the storage manager has time to be created
	time.Sleep(30 * time.Second)

	storageManagers := v2.ListSMs(t, "?all=false", e2e.PowerUser)
	e2e.MustNotEmpty(t, storageManagers, "storage managers were empty")

	storageManager := storageManagers[0]
	assert.NotEmpty(t, storageManager.ID, "storage manager id was empty")
	assert.NotEmpty(t, storageManager.OwnerID, "storage manager owner id was empty")

	v2.WaitForSmRunning(t, storageManager.ID, func(storageManagerRead *body.SmRead) bool {
		// Make sure it is accessible
		if storageManager.URL != nil {
			return v2.CheckUpURL(t, *storageManager.URL)
		}
		return false
	})

	// Ensure the User has the storage url
	user := v2.GetUser(t, model.TestPowerUserID, e2e.PowerUser)
	assert.NotEmpty(t, user.StorageURL, "storage url was empty")
}

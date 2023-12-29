package users

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/test/e2e"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	e2e.Setup()
	code := m.Run()
	e2e.Shutdown()
	os.Exit(code)
}

func TestGet(t *testing.T) {
	t.Parallel()

	e2e.GetUser(t, e2e.AdminUserID)
}

func TestList(t *testing.T) {
	t.Parallel()

	queries := []string{
		// all
		"?all=true",
		// search
		"?search=tester",
	}

	for _, query := range queries {
		users := e2e.ListUsers(t, query)
		assert.NotEmpty(t, users, "users were not fetched. it should have at least one user")
	}
}

func TestDiscover(t *testing.T) {
	t.Parallel()

	queries := []string{
		// search
		"?search=tester&page=1&pageSize=3",
	}

	for _, query := range queries {
		users := e2e.ListUsersDiscovery(t, query)
		assert.NotEmpty(t, users, "users were not fetched. it should have at least one user")
	}
}

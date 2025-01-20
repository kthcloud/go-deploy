package users

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

func TestGetAnotherUser(t *testing.T) {
	t.Parallel()

	// Getting another user requires admin permissions
	v2.GetUser(t, model.TestPowerUserID, e2e.AdminUser)
}

func TestGetYourself(t *testing.T) {
	t.Parallel()

	v2.GetUser(t, model.TestPowerUserID, e2e.PowerUser)
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
		users := v2.ListUsers(t, query)
		assert.NotEmpty(t, users, "users were not fetched. it should have at least one user")
	}
}

func TestDiscover(t *testing.T) {
	t.Parallel()

	queries := []string{
		// Search for the Power user
		"?search=power&page=0&pageSize=3",
	}

	for _, query := range queries {
		users := v2.ListUsersDiscovery(t, query, e2e.DefaultUser)
		assert.NotEmpty(t, users, "users were not fetched. it should have at least one user")
	}
}

func TestDiscoverByID(t *testing.T) {
	t.Parallel()

	user := v2.GetUserDiscovery(t, model.TestPowerUserID, e2e.DefaultUser)
	assert.NotEmpty(t, user, "user was not fetched")
	assert.Equal(t, model.TestPowerUserID, user.ID, "user id does not match")
}

func TestUpdatePublicKeys(t *testing.T) {
	// Since public keys are updated as a whole, we can't run this in parallel

	publicKeys := []body.PublicKey{
		{
			Name: "test-key",
			Key:  "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCjWQF3Wz/DEKVZ+0pTBBGi5XZFWjz3WURwUf9/7zdl/KNO1UNHoaUm6nox0FLFygeI0H1wHsVXYs2L/lYOk9dCerNTWDmxSrvG0hIrwXxrg+xEFoCfOdVX/ItmWkWvIHH4Nk+AnCfO1KacISqWWOX702P0EvEN3E4fTNQmOOJO36VoWk+Hd81/DTJ9ahUhslQWJhGgsUtTIDPdeoL8KuwaQYucBSJrSHK57MXf0REuvybTNL88PX02g02z8du8dV5Sje+7soQY1TblBkAdU15IEYwEd6p8m3/r8ZU56LLp4yG+GvFZBh0HNm7W/3V119fo2qivjM/3JpxR7zoigEHy7AH7gBbtlCjlHIcwHKtaWbTk+J8JvjVv2tI7ug/7C4r224mOx7K/qbOoTUjvRgVKK5jrwSz8EBm0Q12JN0Un6Nf3vw3w0dONYZUVaVnDC49LwIdSBlVYghMVCn7jveN2pe4Sox1DbqffYFsg8HkJnK478+aiNemLLWyL7wEFy90= test-key",
		},
		{
			Name: "test-key-2",
			Key:  "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCjWQF3Wz/DEKVZ+0pTBBGi5XZFWjz3WURwUf9/7zdl/KNO1UNHoaUm6nox0FLFygeI0H1wHsVXYs2L/lYOk9dCerNTWDmxSrvG0hIrwXxrg+xEFoCfOdVX/ItmWkWvIHH4Nk+AnCfO1KacISqWWOX702P0EvEN3E4fTNQmOOJO36VoWk+Hd81/DTJ9ahUhslQWJhGgsUtTIDPdeoL8KuwaQYucBSJrSHK57MXf0REuvybTNL88PX02g02z8du8dV5Sje+7soQY1TblBkAdU15IEYwEd6p8m3/r8ZU56LLp4yG+GvFZBh0HNm7W/3V119fo2qivjM/3JpxR7zoigEHy7AH7gBbtlCjlHIcwHKtaWbTk+J8JvjVv2tI7ug/7C4r224mOx7K/qbOoTUjvRgVKK5jrwSz8EBm0Q12JN0Un6Nf3vw3w0dONYZUVaVnDC49LwIdSBlVYghMVCn7jveN2pe4Sox1DbqffYFsg8HkJnK478+aiNemLLWyL7wEFy90= test-key-2",
		},
	}

	update := body.UserUpdate{
		PublicKeys: &publicKeys,
	}

	v2.UpdateUser(t, model.TestAdminUserID, update)
	v2.UpdateUser(t, model.TestPowerUserID, update)
	v2.UpdateUser(t, model.TestDefaultUserID, update)
}

func TestUpdateUserData(t *testing.T) {
	// Since user data is updated as a whole, we can't run this in parallel

	userData := []body.UserData{
		{
			Key:   "test-key-1",
			Value: "test-value-1",
		},
		{
			Key:   "test-key-2",
			Value: "test-value-2",
		},
	}

	update := body.UserUpdate{
		UserData: &userData,
	}

	v2.UpdateUser(t, model.TestAdminUserID, update)
	v2.UpdateUser(t, model.TestPowerUserID, update)
	v2.UpdateUser(t, model.TestDefaultUserID, update)
}

func TestCreateApiKey(t *testing.T) {
	// Since this edit the user's API keys, we can't run this in parallel

	t.Cleanup(func() {
		v2.UpdateUser(t, model.TestPowerUserID, body.UserUpdate{
			ApiKeys: &[]body.ApiKey{
				{
					Name: e2e.PowerUser,
				},
			},
		})
	})

	names := []string{
		e2e.GenName("test-key-1"),
		e2e.GenName("test-key-2"),
		e2e.GenName("test-key-3"),
		e2e.GenName("test-key-4"),
		e2e.GenName("test-key-5"),
	}

	for _, name := range names {
		v2.CreateApiKey(t, model.TestPowerUserID, body.ApiKeyCreate{
			Name:      name,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		})
	}

	user := v2.GetUser(t, model.TestPowerUserID)
	for _, name := range names {
		found := false
		for _, apiKey := range user.ApiKeys {
			if apiKey.Name == name {
				found = true
				break
			}
		}
		assert.True(t, found, "api key was not created")
	}

	// 1. Clean up the first two one by one
	namesWithoutFirstTwo := names[2:]
	apiKeysWithoutFirstTwo := user.ApiKeys[2:]

	// Add PowerUser to the list to ensure it is not deleted
	apiKeysWithoutFirstTwo = append(apiKeysWithoutFirstTwo, body.ApiKey{
		Name: e2e.PowerUser,
	})

	user = v2.UpdateUser(t, model.TestPowerUserID, body.UserUpdate{
		ApiKeys: &apiKeysWithoutFirstTwo,
	})

	for _, name := range namesWithoutFirstTwo {
		found := false
		for _, apiKey := range user.ApiKeys {
			if apiKey.Name == name {
				found = true
				break
			}
		}
		assert.True(t, found, "api key was not updated")
	}

	// 2. Clean up the rest
	emptyApiKeys := make([]body.ApiKey, 0)

	// Add PowerUser to the list to ensure it is not deleted
	emptyApiKeys = append(emptyApiKeys, body.ApiKey{
		Name: e2e.PowerUser,
	})

	user = v2.UpdateUser(t, model.TestPowerUserID, body.UserUpdate{
		ApiKeys: &emptyApiKeys,
	})
}

func TestMalformedApiKeyDelete(t *testing.T) {
	// Since this edit the user's API keys, we can't run this in parallel

	t.Cleanup(func() {
		v2.UpdateUser(t, model.TestPowerUserID, body.UserUpdate{
			ApiKeys: &[]body.ApiKey{
				{
					Name: e2e.PowerUser,
				},
			},
		})
	})

	names := []string{
		e2e.GenName("test-key-1"),
		e2e.GenName("test-key-2"),
		e2e.GenName("test-key-3"),
	}

	for _, name := range names {
		v2.CreateApiKey(t, model.TestPowerUserID, body.ApiKeyCreate{
			Name:      name,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		})
	}

	user := v2.GetUser(t, model.TestPowerUserID)
	for _, name := range names {
		found := false
		for _, apiKey := range user.ApiKeys {
			if apiKey.Name == name {
				found = true
				break
			}
		}
		assert.True(t, found, "api key was not created")
	}

	// Update with test-key-1, test-key-2 and unknown-key
	// This should skip the unknown-key and delete test-key-3
	culprit := e2e.GenName("unknown-key")
	newKeys := []body.ApiKey{
		{
			Name: e2e.PowerUser,
		},
		{
			Name: names[0],
		},
		{
			Name: names[1],
		},
		{
			Name: culprit,
		},
	}

	user = v2.UpdateUser(t, model.TestPowerUserID, body.UserUpdate{
		ApiKeys: &newKeys,
	})

	nameThatShouldExist := names[:2]

	// Add PowerUser to the list to ensure it is not deleted
	nameThatShouldExist = append(nameThatShouldExist, e2e.PowerUser)

	for _, name := range nameThatShouldExist {
		found := false
		for _, apiKey := range user.ApiKeys {
			if apiKey.Name == name {
				found = true
				break
			}
		}
		assert.True(t, found, "api key was not updated")
	}

	// Ensure culprit key is not added to the list
	for _, apiKey := range user.ApiKeys {
		assert.NotEqual(t, culprit, apiKey.Name, "unknown-key was added")
		assert.NotEmpty(t, apiKey.Name, "empty key was added")
		assert.NotEmpty(t, apiKey.CreatedAt, "empty created at was added")
		assert.NotEmpty(t, apiKey.ExpiresAt, "empty created at was added")
	}
}

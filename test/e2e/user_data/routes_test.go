package user_data

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/dto/v1/body"
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

	userData := e2e.WithUserData(t, body.UserDataCreate{
		ID:   e2e.GenName("userData"),
		Data: "test",
	})

	userDataRead := e2e.GetUserData(t, userData.ID)
	if userDataRead.ID != userData.ID {
		t.Error("user data was not fetched")
	}
}

func TestList(t *testing.T) {
	t.Parallel()

	userData := e2e.WithUserData(t, body.UserDataCreate{
		ID:   e2e.GenName("userData"),
		Data: "test",
	})

	userDataList := e2e.ListUserData(t, "")
	assert.NotEmpty(t, userDataList, "user data was not fetched")

	found := false
	for _, ud := range userDataList {
		if ud.ID == userData.ID {
			found = true
			break
		}
	}

	assert.True(t, found, "user data was not fetched")
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	userData := e2e.WithUserData(t, body.UserDataCreate{
		ID:   e2e.GenName("userData"),
		Data: "test",
	})

	userDataUpdate := e2e.UpdateUserData(t, userData.ID, body.UserDataUpdate{
		Data: "test2",
	})

	assert.Equal(t, "test2", userDataUpdate.Data, "user data was not updated")
}

package notifications

import (
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/test/e2e"
	"github.com/kthcloud/go-deploy/test/e2e/v2"
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
	queries := []string{
		"?page=1&pageSize=10",
	}

	for _, query := range queries {
		v2.ListNotifications(t, query)
	}
}

func TestMarkRead(t *testing.T) {
	// We need to get a notification to mark it as read
	// We can create a team and invite a user to get a notification

	// Create a team and invite a user
	v2.WithTeam(t, body.TeamCreate{
		Name:        e2e.GenName(),
		Description: e2e.GenName(),
		Resources:   nil,
		Members: []body.TeamMemberCreate{
			{
				ID: model.TestDefaultUserID,
			},
		},
	}, e2e.PowerUser)

	// List notifications for the user
	notifications := v2.ListNotifications(t, "?page=0&pageSize=100", e2e.DefaultUser)
	if len(notifications) == 0 {
		t.Fatal("no notifications found")
	}

	// Mark a notification as read
	v2.UpdateNotification(t, notifications[0].ID, body.NotificationUpdate{
		Read: true,
	}, e2e.DefaultUser)
}

func TestMarkToasted(t *testing.T) {
	// We need to get a notification to mark it as toasted
	// We can create a team and invite a user to get a notification

	// Create a team and invite a user
	v2.WithTeam(t, body.TeamCreate{
		Name:        e2e.GenName(),
		Description: e2e.GenName(),
		Resources:   nil,
		Members: []body.TeamMemberCreate{
			{
				ID: model.TestDefaultUserID,
			},
		},
	}, e2e.PowerUser)

	// List notifications for the user
	notifications := v2.ListNotifications(t, "?page=0&pageSize=100", e2e.DefaultUser)
	if len(notifications) == 0 {
		t.Fatal("no notifications found")
	}

	// Mark a notification as toasted
	v2.UpdateNotification(t, notifications[0].ID, body.NotificationUpdate{
		Toasted: true,
	}, e2e.DefaultUser)
}

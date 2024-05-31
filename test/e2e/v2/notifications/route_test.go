package notifications

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
		v2.ListNotifications(t, query)
	}
}

func TestMarkRead(t *testing.T) {
	// We need to get a notification to mark it as read
	// We can create a team and invite a user to get a notification

	justBefore := time.Now()

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
	assert.NotEmpty(t, notifications, "no notifications found")

	justAfter := time.Now()

	// Get the notification
	var notification body.NotificationRead
	for _, n := range notifications {
		if n.Type == model.NotificationTeamInvite && n.CreatedAt.After(justBefore) && n.CreatedAt.Before(justAfter) {
			notification = n
			break
		}
	}

	// Mark the notification as read
	v2.UpdateNotification(t, notification.ID, body.NotificationUpdate{
		Read: true,
	}, e2e.DefaultUser)
}

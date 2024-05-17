package notifications

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/dto/v1/body"
	"go-deploy/models/model"
	"go-deploy/test/e2e"
	v1 "go-deploy/test/e2e/v1"
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
		v1.ListNotifications(t, query)
	}
}

func TestMarkRead(t *testing.T) {
	// We need to get a notification to mark it as read
	// We can create a team and invite a user to get a notification

	justBefore := time.Now()

	// Create a team and invite a user
	v1.WithTeam(t, body.TeamCreate{
		Name:        "",
		Description: "",
		Resources:   nil,
		Members: []body.TeamMemberCreate{
			{
				ID: e2e.DefaultUserID,
			},
		},
	}, e2e.PowerUserID)

	// List notifications for the user
	notifications := v1.ListNotifications(t, "?page=1&pageSize=10", e2e.DefaultUserID)
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
	v1.UpdateNotification(t, notification.ID, body.NotificationUpdate{
		Read: true,
	}, e2e.DefaultUserID)
}

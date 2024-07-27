package v2

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/dto/v2/body"
	"go-deploy/test/e2e"
	"testing"
)

const (
	NotificationPath  = "/v2/notifications/"
	NotificationsPath = "/v2/notifications"
)

func GetNotification(t *testing.T, id string, user ...string) body.NotificationRead {
	resp := e2e.DoGetRequest(t, NotificationPath+id, user...)
	return e2e.MustParse[body.NotificationRead](t, resp)
}

func ListNotifications(t *testing.T, query string, user ...string) []body.NotificationRead {
	resp := e2e.DoGetRequest(t, NotificationsPath+query, user...)
	return e2e.MustParse[[]body.NotificationRead](t, resp)
}

func UpdateNotification(t *testing.T, id string, update body.NotificationUpdate, user ...string) body.NotificationRead {
	resp := e2e.DoPostRequest(t, NotificationPath+id, update, user...)
	updatedNotification := e2e.MustParse[body.NotificationRead](t, resp)

	if update.Read {
		assert.NotNil(t, updatedNotification.ReadAt)
	}

	if update.Toasted {
		assert.NotNil(t, updatedNotification.ToastedAt)
	}

	return updatedNotification
}

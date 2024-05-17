package v1

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/dto/v1/body"
	"go-deploy/test/e2e"
	"testing"
)

const (
	NotificationPath  = "/v1/notifications/"
	NotificationsPath = "/v1/notifications"
)

func GetNotification(t *testing.T, id string, userID ...string) body.NotificationRead {
	resp := e2e.DoGetRequest(t, NotificationPath+id, userID...)
	return e2e.MustParse[body.NotificationRead](t, resp)
}

func ListNotifications(t *testing.T, query string, userID ...string) []body.NotificationRead {
	resp := e2e.DoGetRequest(t, NotificationsPath+query, userID...)
	return e2e.MustParse[[]body.NotificationRead](t, resp)
}

func UpdateNotification(t *testing.T, id string, update body.NotificationUpdate, userID ...string) body.NotificationRead {
	resp := e2e.DoPostRequest(t, NotificationPath+id, update, userID...)
	updatedNotification := e2e.MustParse[body.NotificationRead](t, resp)

	if update.Read {
		assert.NotNil(t, updatedNotification.ReadAt)
	}

	return updatedNotification
}

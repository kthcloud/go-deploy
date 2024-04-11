package v1

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/dto/v1/body"
	"go-deploy/test/e2e"
	"net/http"
	"testing"
)

const (
	NotificationPath  = "/v1/notifications/"
	NotificationsPath = "/v1/notifications"
)

func GetNotification(t *testing.T, id string, userID ...string) body.NotificationRead {
	resp := e2e.DoGetRequest(t, NotificationPath+id, userID...)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "notification was not fetched")

	var notificationRead body.NotificationRead
	err := e2e.ReadResponseBody(t, resp, &notificationRead)
	assert.NoError(t, err, "notification was not fetched")

	return notificationRead
}

func ListNotifications(t *testing.T, query string, userID ...string) []body.NotificationRead {
	resp := e2e.DoGetRequest(t, NotificationsPath+query, userID...)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "notifications were not fetched")

	var notifications []body.NotificationRead
	err := e2e.ReadResponseBody(t, resp, &notifications)
	assert.NoError(t, err, "notifications were not fetched")

	return notifications
}

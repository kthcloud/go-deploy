package e2e

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/body"
	"net/http"
	"testing"
)

func GetNotification(t *testing.T, id string, userID ...string) body.NotificationRead {
	resp := DoGetRequest(t, "/notifications/"+id, userID...)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "notification was not fetched")

	var notificationRead body.NotificationRead
	err := ReadResponseBody(t, resp, &notificationRead)
	assert.NoError(t, err, "notification was not fetched")

	return notificationRead
}

func ListNotifications(t *testing.T, query string, userID ...string) []body.NotificationRead {
	resp := DoGetRequest(t, "/notifications"+query, userID...)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "notifications were not fetched")

	var notifications []body.NotificationRead
	err := ReadResponseBody(t, resp, &notifications)
	assert.NoError(t, err, "notifications were not fetched")

	return notifications
}

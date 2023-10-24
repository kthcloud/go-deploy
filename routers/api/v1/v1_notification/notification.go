package v1_notification

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/models/dto/uri"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/notification_service"
	"net/http"
)

// Get godoc
// @Summary Get notification
// @Description Get notification
// @Tags Notification
// @Accept  json
// @Produce  json
// @Param notificationId path string true "Notification ID"
// @Success 200 {object} body.NotificationRead
// @Failure 400 {object} sys.ErrorResponse
// @Router /notifications/{notificationId} [get]
func Get(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery uri.NotificationUpdate
	if err := context.GinContext.ShouldBindUri(&requestQuery); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	var requestBody body.NotificationUpdate
	if err := context.GinContext.ShouldBindJSON(&requestBody); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	notification, err := notification_service.GetByIdAuth(requestQuery.NotificationID, auth)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get notification: %s", err))
		return
	}

	context.JSONResponse(http.StatusOK, notification.ToDTO())
}

// GetList godoc
// @Summary Get notifications
// @Description Get notifications
// @Tags Notification
// @Accept  json
// @Produce  json
// @Param all query bool false "Get all notifications"
// @Param userId query string false "Get notifications by user id"
// @Param page query int false "Page number"
// @Param pageSize query int false "Number of items per page"
// @Success 200 {array} body.NotificationRead
// @Failure 400 {object} sys.ErrorResponse
// @Router /notifications [get]
func GetList(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.NotificationList
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	notifications, err := notification_service.GetManyAuth(requestQuery.All, requestQuery.UserID, auth, &requestQuery.Pagination)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get notifications: %s", err))
		return
	}

	dtoNotifications := make([]body.NotificationRead, len(notifications))
	for i, notification := range notifications {
		dtoNotifications[i] = notification.ToDTO()
	}

	context.JSONResponse(http.StatusOK, dtoNotifications)
}

// Update godoc
// @Summary Update notification
// @Description Update notification
// @Tags Notification
// @Accept  json
// @Produce  json
// @Param notificationId path string true "Notification ID"
// @Param body body body.NotificationUpdate true "Notification update"
// @Success 200
// @Failure 400 {object} sys.ErrorResponse
// @Router /notifications/{notificationId} [post]
func Update(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery uri.NotificationUpdate
	if err := context.GinContext.ShouldBindUri(&requestQuery); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	var requestBody body.NotificationUpdate
	if err := context.GinContext.ShouldBindJSON(&requestBody); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	err = notification_service.UpdateAuth(requestQuery.NotificationID, &requestBody, auth)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to update notification: %s", err))
		return
	}

	context.Ok()
}

// Delete godoc
// @Summary Delete notification
// @Description Delete notification
// @Tags Notification
// @Accept  json
// @Produce  json
// @Param notificationId path string true "Notification ID"
// @Success 200
// @Failure 400 {object} sys.ErrorResponse
// @Router /notifications/{notificationId} [delete]
func Delete(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery uri.NotificationDelete
	if err := context.GinContext.ShouldBindUri(&requestQuery); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	err = notification_service.DeleteAuth(requestQuery.NotificationID, auth)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to delete notification: %s", err))
		return
	}

	context.OkDeleted()
}

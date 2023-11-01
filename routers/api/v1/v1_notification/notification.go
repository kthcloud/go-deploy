package v1_notification

import (
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/models/dto/uri"
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

	var requestQuery uri.NotificationGet
	if err := context.GinContext.ShouldBindUri(&requestQuery); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	notification, err := notification_service.GetByIdWithAuth(requestQuery.NotificationID, auth)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	context.JSONResponse(http.StatusOK, notification.ToDTO())
}

// List godoc
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
func List(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.NotificationList
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	notifications, err := notification_service.ListAuth(requestQuery.All, requestQuery.UserID, auth, &requestQuery.Pagination)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	dtoNotifications := make([]body.NotificationRead, len(notifications))
	for i, notification := range notifications {
		dtoNotifications[i] = notification.ToDTO()
	}

	context.Ok(dtoNotifications)
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
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	var requestBody body.NotificationUpdate
	if err := context.GinContext.ShouldBindJSON(&requestBody); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	updated, err := notification_service.UpdateAuth(requestQuery.NotificationID, &requestBody, auth)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if updated == nil {
		context.NotFound("Notification not found")
		return
	}

	context.Ok(updated.ToDTO())
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
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	err = notification_service.DeleteAuth(requestQuery.NotificationID, auth)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	context.OkNoContent()
}

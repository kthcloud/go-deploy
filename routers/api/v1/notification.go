package v1

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go-deploy/dto/v1/body"
	"go-deploy/dto/v1/query"
	"go-deploy/dto/v1/uri"
	"go-deploy/pkg/sys"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/v1/notifications/opts"
	v12 "go-deploy/service/v1/utils"
	"net/http"
)

// GetNotification godoc
// @Summary Get notification
// @Description Get notification
// @Tags Notification
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param notificationId path string true "Notification ID"
// @Success 200 {object} body.NotificationRead
// @Failure 400 {object} sys.ErrorResponse
// @Router /v1/notifications/{notificationId} [get]
func GetNotification(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery uri.NotificationGet
	if err := context.GinContext.ShouldBindUri(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	notification, err := service.V1(auth).Notifications().Get(requestQuery.NotificationID)
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	context.JSONResponse(http.StatusOK, notification.ToDTO())
}

// ListNotifications godoc
// @Summary List notifications
// @Description List notifications
// @Tags Notification
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param all query bool false "List all notifications"
// @Param userId query string false "Filter by user ID"
// @Param page query int false "Page number"
// @Param pageSize query int false "Number of items per page"
// @Success 200 {array} body.NotificationRead
// @Failure 400 {object} sys.ErrorResponse
// @Router /v1/notifications [get]
func ListNotifications(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.NotificationList
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	var userID *string
	if requestQuery.UserID != nil {
		userID = requestQuery.UserID
	} else if !requestQuery.All {
		userID = &auth.User.ID
	}

	notificationList, err := service.V1(auth).Notifications().List(opts.ListOpts{
		Pagination: v12.GetOrDefaultPagination(requestQuery.Pagination),
		UserID:     userID,
	})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	dtoNotifications := make([]body.NotificationRead, len(notificationList))
	for i, notification := range notificationList {
		dtoNotifications[i] = notification.ToDTO()
	}

	context.Ok(dtoNotifications)
}

// UpdateNotification godoc
// @Summary Update notification
// @Description Update notification
// @Tags Notification
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param notificationId path string true "Notification ID"
// @Param body body body.NotificationUpdate true "Notification update"
// @Success 200
// @Failure 400 {object} sys.ErrorResponse
// @Router /v1/notifications/{notificationId} [post]
func UpdateNotification(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery uri.NotificationUpdate
	if err := context.GinContext.ShouldBindUri(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	var requestBody body.NotificationUpdate
	if err := context.GinContext.ShouldBindJSON(&requestBody); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	updated, err := service.V1(auth).Notifications().Update(requestQuery.NotificationID, &requestBody)
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if updated == nil {
		context.NotFound("Notification not found")
		return
	}

	context.Ok(updated.ToDTO())
}

// DeleteNotification godoc
// @Summary Delete notification
// @Description Delete notification
// @Tags Notification
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param notificationId path string true "Notification ID"
// @Success 200
// @Failure 400 {object} sys.ErrorResponse
// @Router /v1/notifications/{notificationId} [delete]
func DeleteNotification(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery uri.NotificationDelete
	if err := context.GinContext.ShouldBindUri(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	err = service.V1(auth).Notifications().Delete(requestQuery.NotificationID)
	if err != nil {
		if errors.Is(err, sErrors.NotificationNotFoundErr) {
			context.NotFound("Notification not found")
			return
		}

		context.ServerError(err, InternalError)
		return
	}

	context.OkNoContent()
}

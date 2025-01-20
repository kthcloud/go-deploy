package sys

import (
	"github.com/gin-gonic/gin"
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/app/status_codes"
	"github.com/kthcloud/go-deploy/utils"
	"net/http"
)

// ClientContext is a wrapper for the gin context.
type ClientContext struct {
	GinContext *gin.Context
}

// NewContext creates a new client context.
func NewContext(ginContext *gin.Context) ClientContext {
	return ClientContext{GinContext: ginContext}
}

type ErrorResponse struct {
	Errors []Error `json:"errors"`
}

type Error struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

type validationErrorResponse struct {
	ValidationErrors map[string][]string `json:"validationErrors"`
}

// GetAuthUser gets the authenticated user from the context.
func (context *ClientContext) GetAuthUser() *model.User {
	if val, exists := context.GinContext.Get("authUser"); exists && val != nil {
		asUser, ok := val.(*model.User)
		if ok {
			return asUser
		}
	}

	return nil
}

// ResponseValidationError is a helper function to return a validation error response.
func (context *ClientContext) ResponseValidationError(errors map[string][]string) {
	context.GinContext.JSON(400, validationErrorResponse{ValidationErrors: errors})
}

// ServerError is a helper function to return a server error response.
func (context *ClientContext) ServerError(log, display error) {
	utils.PrettyPrintError(log)
	context.GinContext.JSON(http.StatusInternalServerError, ErrorResponse{Errors: []Error{{Code: status_codes.GetMsg(status_codes.Error), Msg: display.Error()}}})
}

// ServerUnavailableError is a helper function to return a server unavailable error response.
// It logs the error internally, and returns a generic error message to the user.
func (context *ClientContext) ServerUnavailableError(log, display error) {
	utils.PrettyPrintError(log)
	context.GinContext.JSON(http.StatusServiceUnavailable, ErrorResponse{Errors: []Error{{Code: status_codes.GetMsg(status_codes.Error), Msg: display.Error()}}})
}

// UserError is a helper function to return a user error response.
func (context *ClientContext) UserError(msg string) {
	context.GinContext.JSON(http.StatusBadRequest, ErrorResponse{Errors: []Error{{Code: status_codes.GetMsg(status_codes.Error), Msg: msg}}})
}

// BindingError is a helper function to return a binding error response.
// This normally occurs when the user sends a request with invalid data.
func (context *ClientContext) BindingError(bindingError *body.BindingError) {
	context.GinContext.JSON(http.StatusBadRequest, bindingError)
}

// ErrorResponse is a helper function to return an error response.
func (context *ClientContext) ErrorResponse(httpCode int, errCode int, message string) {
	errors := []Error{{Code: status_codes.GetMsg(errCode), Msg: message}}
	context.GinContext.JSON(httpCode, ErrorResponse{Errors: errors})
}

// JSONResponse is a helper function to return a JSON response.
func (context *ClientContext) JSONResponse(httpCode int, data interface{}) {
	context.GinContext.JSON(httpCode, data)
}

// Ok is a helper function to return an OK response.
func (context *ClientContext) Ok(data interface{}) {
	context.GinContext.JSON(http.StatusOK, data)
}

// OkNoContent is a helper function to return an OK response with no content.
func (context *ClientContext) OkNoContent() {
	context.GinContext.Status(http.StatusNoContent)
}

// Unauthorized is a helper function to return an unauthorized response.
func (context *ClientContext) Unauthorized(msg string) {
	context.GinContext.JSON(http.StatusUnauthorized, ErrorResponse{Errors: []Error{{Code: status_codes.GetMsg(status_codes.Error), Msg: msg}}})
}

// Forbidden is a helper function to return a forbidden response.
func (context *ClientContext) Forbidden(msg string) {
	context.GinContext.JSON(http.StatusForbidden, ErrorResponse{Errors: []Error{{Code: status_codes.GetMsg(status_codes.Error), Msg: msg}}})
}

// NotFound is a helper function to return a not found response.
func (context *ClientContext) NotFound(msg string) {
	context.GinContext.JSON(http.StatusNotFound, ErrorResponse{Errors: []Error{{Code: status_codes.GetMsg(status_codes.Error), Msg: msg}}})
}

// Locked is a helper function to return a locked response.
func (context *ClientContext) Locked(msg string) {
	context.GinContext.JSON(http.StatusLocked, ErrorResponse{Errors: []Error{{Code: status_codes.GetMsg(status_codes.Error), Msg: msg}}})
}

// NotModified is a helper function to return a not modified response.
func (context *ClientContext) NotModified() {
	context.GinContext.Status(http.StatusNotModified)
}

// NotImplemented is a helper function to return a not implemented response.
func (context *ClientContext) NotImplemented() {
	context.GinContext.Status(http.StatusNotImplemented)
}

// OkCreated is a helper function to return a created response.
func (context *ClientContext) OkCreated(data interface{}) {
	context.GinContext.JSON(http.StatusCreated, data)
}

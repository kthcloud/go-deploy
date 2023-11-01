package sys

import (
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/body"
	"go-deploy/pkg/app/status_codes"
	"go-deploy/utils"
	"net/http"
)

type ClientContext struct {
	GinContext *gin.Context
}

func NewContext(ginContext *gin.Context) ClientContext {
	return ClientContext{GinContext: ginContext}
}

type ErrorResponse struct {
	Errors []errorPiece `json:"errors"`
}

type errorPiece struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

type validationErrorResponse struct {
	ValidationErrors map[string][]string `json:"validationErrors"`
}

func (context *ClientContext) ResponseValidationError(errors map[string][]string) {
	context.GinContext.JSON(400, validationErrorResponse{ValidationErrors: errors})
}

func (context *ClientContext) ServerError(log, display error) {
	utils.PrettyPrintError(log)
	context.GinContext.JSON(http.StatusInternalServerError, ErrorResponse{Errors: []errorPiece{{Code: status_codes.GetMsg(status_codes.Error), Msg: display.Error()}}})
}

func (context *ClientContext) ServerUnavailableError(log, display error) {
	utils.PrettyPrintError(log)
	context.GinContext.JSON(http.StatusServiceUnavailable, ErrorResponse{Errors: []errorPiece{{Code: status_codes.GetMsg(status_codes.Error), Msg: display.Error()}}})
}

func (context *ClientContext) UserError(msg string) {
	context.GinContext.JSON(http.StatusBadRequest, ErrorResponse{Errors: []errorPiece{{Code: status_codes.GetMsg(status_codes.Error), Msg: msg}}})
}

func (context *ClientContext) BindingError(bindingError *body.BindingError) {
	context.GinContext.JSON(http.StatusBadRequest, bindingError)
}

func (context *ClientContext) ErrorResponse(httpCode int, errCode int, message string) {
	errors := []errorPiece{{Code: status_codes.GetMsg(errCode), Msg: message}}
	context.GinContext.JSON(httpCode, ErrorResponse{Errors: errors})
}

func (context *ClientContext) JSONResponse(httpCode int, data interface{}) {
	context.GinContext.JSON(httpCode, data)
}

func (context *ClientContext) Ok(data interface{}) {
	context.GinContext.JSON(http.StatusOK, data)
}

func (context *ClientContext) OkNoContent() {
	context.GinContext.Status(http.StatusNoContent)
}

func (context *ClientContext) Unauthorized(msg string) {
	context.GinContext.JSON(http.StatusUnauthorized, ErrorResponse{Errors: []errorPiece{{Code: status_codes.GetMsg(status_codes.Error), Msg: msg}}})
}

func (context *ClientContext) Forbidden(msg string) {
	context.GinContext.JSON(http.StatusForbidden, ErrorResponse{Errors: []errorPiece{{Code: status_codes.GetMsg(status_codes.Error), Msg: msg}}})
}

func (context *ClientContext) NotFound(msg string) {
	context.GinContext.JSON(http.StatusNotFound, ErrorResponse{Errors: []errorPiece{{Code: status_codes.GetMsg(status_codes.Error), Msg: msg}}})
}

func (context *ClientContext) Locked(msg string) {
	context.GinContext.JSON(http.StatusLocked, ErrorResponse{Errors: []errorPiece{{Code: status_codes.GetMsg(status_codes.Error), Msg: msg}}})
}

func (context *ClientContext) NotModified() {
	context.GinContext.Status(http.StatusNotModified)
}

func (context *ClientContext) OkCreated(data interface{}) {
	context.GinContext.JSON(http.StatusCreated, data)
}

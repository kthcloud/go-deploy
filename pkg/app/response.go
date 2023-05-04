package app

import (
	deployApiErrors "go-deploy/pkg/status_codes"
	"net/http"
)

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

func (context *ClientContext) ErrorResponse(httpCode int, errCode int, message string) {
	errors := []errorPiece{{Code: deployApiErrors.GetMsg(errCode), Msg: message}}
	context.GinContext.JSON(httpCode, ErrorResponse{Errors: errors})
}

func (context *ClientContext) JSONResponse(httpCode int, data interface{}) {
	context.GinContext.JSON(httpCode, data)
}

func (context *ClientContext) Ok() {
	context.GinContext.Status(http.StatusOK)
}

func (context *ClientContext) OkDeleted() {
	context.GinContext.Status(http.StatusNoContent)
}

func (context *ClientContext) Unauthorized() {
	context.GinContext.Status(http.StatusUnauthorized)
}

func (context *ClientContext) NotFound() {
	context.GinContext.Status(http.StatusNotFound)
}

func (context *ClientContext) NotModified() {
	context.GinContext.Status(http.StatusNotModified)
}

func (context *ClientContext) OkCreated() {
	context.GinContext.Status(http.StatusCreated)
}

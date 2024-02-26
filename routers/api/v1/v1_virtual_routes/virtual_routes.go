package v1_virtual_routes

import "github.com/gin-gonic/gin"

// Virtual routes represent routes that are not handled by go-deploy, but some external package.
// These packages might not support docs that enables swagger to generate documentation for them.
// So we need to manually add them here.

// Metrics
// @Summary Get metrics
// @Description Get metrics
// @Tags Metrics
// @Accept  json
// @Produce  json
// @Success 200 {object}  string
// @Failure 500 {object} sys.ErrorResponse
// @Router /metrics [get]
func Metrics(c *gin.Context) {}

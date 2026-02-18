package v2

import "github.com/gin-gonic/gin"

// Virtual routes represent routes that are not handled by go-deploy, but some external package.
// These packages might not support docs that enables swagger to generate documentation for them.
// So we need to manually add them here.

// Metrics
// @Summary Get metrics
// @Description Get metrics
// @Tags Metrics
// @Produce json
// @Success 200 {object}  string
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/metrics [get]
func Metrics(c *gin.Context) {}

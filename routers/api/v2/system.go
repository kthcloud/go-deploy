package v2

import (
	"github.com/gin-gonic/gin"
	"github.com/kthcloud/go-deploy/dto/v2/query"
	"github.com/kthcloud/go-deploy/pkg/sys"
	"github.com/kthcloud/go-deploy/service"
	"net/http"
)

// ListSystemCapacities
// @Summary List system capacities
// @Description List system capacities
// @Tags System
// @Accept  json
// @Produce  json
// @Param n query int false "n"
// @Success 200 {array} body.TimestampedSystemCapacities
// @Failure 400 {object} body.BindingError
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/systemCapacities [get]
func ListSystemCapacities(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.TimestampRequest
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.JSONResponse(http.StatusBadRequest, CreateBindingError(err))
		return
	}

	capacities, err := service.V2().System().ListCapacities(requestQuery.N)
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if capacities == nil {
		context.JSONResponse(200, make([]interface{}, 0))
		return
	}

	context.JSONResponse(200, capacities)
}

// ListSystemStats
// @Summary List system stats
// @Description List system stats
// @Tags System
// @Accept  json
// @Produce  json
// @Param n query int false "n"
// @Success 200 {array} body.TimestampedSystemStats
// @Failure 400 {object} body.BindingError
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/systemStats [get]
func ListSystemStats(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.TimestampRequest
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.JSONResponse(http.StatusBadRequest, CreateBindingError(err))
		return
	}

	stats, err := service.V2().System().ListStats(requestQuery.N)
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if stats == nil {
		context.JSONResponse(200, make([]interface{}, 0))
		return
	}

	context.JSONResponse(200, stats)
}

// ListSystemStatus
// @Summary List system stats
// @Description List system stats
// @Tags System
// @Accept  json
// @Produce  json
// @Param n query int false "n"
// @Success 200 {array} body.TimestampedSystemStatus
// @Failure 400 {object} body.BindingError
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/systemStatus [get]
func ListSystemStatus(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.TimestampRequest
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.JSONResponse(http.StatusBadRequest, CreateBindingError(err))
		return
	}

	status, err := service.V2().System().ListStatus(requestQuery.N)
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if status == nil {
		context.JSONResponse(200, make([]interface{}, 0))
		return
	}

	context.JSONResponse(200, status)
}

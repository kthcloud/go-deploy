package v2

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/dto/v2/query"
	"go-deploy/pkg/sys"
	"go-deploy/service"
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
		context.ServerError(err, fmt.Errorf("failed to fetch capacities"))
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
// @Success 200 {array} body.TimestampedSystemCapacities
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

	capacities, err := service.V2().System().ListCapacities(requestQuery.N)
	if err != nil {
		context.ServerError(err, fmt.Errorf("failed to fetch capacities"))
		return
	}

	if capacities == nil {
		context.JSONResponse(200, make([]interface{}, 0))
		return
	}

	context.JSONResponse(200, capacities)
}

// ListSystemStatus
// @Summary List system stats
// @Description List system stats
// @Tags System
// @Accept  json
// @Produce  json
// @Param n query int false "n"
// @Success 200 {array} body.TimestampedSystemCapacities
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

	capacities, err := service.V2().System().ListCapacities(requestQuery.N)
	if err != nil {
		context.ServerError(err, fmt.Errorf("failed to fetch capacities"))
		return
	}

	if capacities == nil {
		context.JSONResponse(200, make([]interface{}, 0))
		return
	}

	context.JSONResponse(200, capacities)
}

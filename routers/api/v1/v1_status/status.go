package v1_status

import (
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/v1/body"
	"go-deploy/models/dto/v1/query"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
	"go-deploy/service/v1/status/opts"
)

// List
// @Summary Get list of zones
// @Description Get list of zones
// @Tags zone
// @Accept json
// @Produce json
// @Param type query string false "Zone type"
// @Success 200 {array} body.ZoneRead
func List(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.StatusList
	if err := context.GinContext.Bind(&requestQuery); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	workerStatus, err := service.V1().Status().ListWorkerStatus(opts.ListWorkerStatusOpts{})
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	dtoWorkerStatus := make([]body.WorkerStatusRead, len(workerStatus))
	for i, zone := range workerStatus {
		dtoWorkerStatus[i] = zone.ToDTO()
	}

	context.Ok(dtoWorkerStatus)
}

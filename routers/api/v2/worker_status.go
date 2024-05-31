package v2

import (
	"github.com/gin-gonic/gin"
	"go-deploy/dto/v2/body"
	"go-deploy/dto/v2/query"
	"go-deploy/pkg/sys"
	"go-deploy/service"
	"go-deploy/service/v2/status/opts"
)

// ListWorkerStatus
// @Summary List worker status
// @Description List of worker status
// @Tags Status
// @Accept json
// @Produce json
// @Success 200 {array} body.WorkerStatusRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/status [get]
func ListWorkerStatus(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.StatusList
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	workerStatus, err := service.V2().Status().ListWorkerStatus(opts.ListWorkerStatusOpts{})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	dtoWorkerStatus := make([]body.WorkerStatusRead, len(workerStatus))
	for i, zone := range workerStatus {
		dtoWorkerStatus[i] = zone.ToDTO()
	}

	context.Ok(dtoWorkerStatus)
}

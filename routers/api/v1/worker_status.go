package v1

import (
	"github.com/gin-gonic/gin"
	"go-deploy/dto/v1/body"
	"go-deploy/dto/v1/query"
	"go-deploy/pkg/sys"
	"go-deploy/service"
	"go-deploy/service/v1/status/opts"
)

// ListWorkerStatus
// @Summary Get list of worker status
// @Description Get list of worker status
// @Tags zone
// @Accept json
// @Produce json
// @Success 200 {array} body.WorkerStatusRead
func ListWorkerStatus(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.StatusList
	if err := context.GinContext.Bind(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	workerStatus, err := service.V1().Status().ListWorkerStatus(opts.ListWorkerStatusOpts{})
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

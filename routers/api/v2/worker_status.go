package v2

import (
	"github.com/gin-gonic/gin"
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/dto/v2/query"
	"github.com/kthcloud/go-deploy/pkg/sys"
	"github.com/kthcloud/go-deploy/service"
	"github.com/kthcloud/go-deploy/service/v2/system/opts"
)

// ListWorkerStatus
// @Summary List worker status
// @Description List of worker status
// @Tags Status
// @Produce json
// @Success 200 {array} body.WorkerStatusRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/workerStatus [get]
func ListWorkerStatus(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.StatusList
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	workerStatus, err := service.V2().System().ListWorkerStatus(opts.ListWorkerStatusOpts{})
	if err != nil {
		context.ServerError(err, ErrInternal)
		return
	}

	dtoWorkerStatus := make([]body.WorkerStatusRead, len(workerStatus))
	for i, zone := range workerStatus {
		dtoWorkerStatus[i] = zone.ToDTO()
	}

	context.Ok(dtoWorkerStatus)
}

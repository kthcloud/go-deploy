package v1_zone

import (
	"github.com/gin-gonic/gin"
	"go-deploy/dto/v1/body"
	"go-deploy/dto/v1/query"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
	"go-deploy/service/v1/zones/opts"
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

	var requestQuery query.ZoneList
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	zoneList, err := service.V1(auth).Zones().List(opts.ListOpts{Type: requestQuery.Type})
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	dtoZones := make([]body.ZoneRead, len(zoneList))
	for i, zone := range zoneList {
		dtoZones[i] = zone.ToDTO()
	}

	context.Ok(dtoZones)
}

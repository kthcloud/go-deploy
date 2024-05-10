package v1

import (
	"github.com/gin-gonic/gin"
	"go-deploy/dto/v1/body"
	"go-deploy/dto/v1/query"
	"go-deploy/pkg/sys"
	"go-deploy/service"
)

// ListZones
// @Summary List zones
// @Description List zones
// @Tags Zone
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} body.ZoneRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v1/zones [get]
func ListZones(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.ZoneList
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	zoneList, err := service.V1(auth).Zones().List()
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	dtoZones := make([]body.ZoneRead, len(zoneList))
	for i, zone := range zoneList {
		dtoZones[i] = zone.ToDTO()
	}

	legacyZoneList, err := service.V1(auth).Zones().ListLegacy()
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	dtoLegacyZones := make([]body.ZoneRead, len(legacyZoneList))
	for i, zone := range legacyZoneList {
		dtoLegacyZones[i] = zone.ToDTO()
	}

	dtoZones = append(dtoZones, dtoLegacyZones...)

	context.Ok(dtoZones)
}

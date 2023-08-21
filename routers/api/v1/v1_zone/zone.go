package v1_zone

import (
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	zoneModel "go-deploy/models/sys/zone"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/zone_service"
	"net/http"
)

// GetList
// @Summary Get list of zones
// @Description Get list of zones
// @Tags zone
// @Accept json
// @Produce json
// @Param type query string false "Zone type"
// @Success 200 {array} body.ZoneRead
func GetList(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.ZoneList
	if err := context.GinContext.Bind(&requestQuery); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	var zones []zoneModel.Zone
	if requestQuery.Type == nil {
		zones, _ = zone_service.GetAllZones()
	} else {
		zones, _ = zone_service.GetZonesByType(*requestQuery.Type)
	}

	dtoZones := make([]body.ZoneRead, len(zones))
	for i, zone := range zones {
		dtoZones[i] = zone.ToDTO()
	}

	context.JSONResponse(http.StatusOK, dtoZones)
}

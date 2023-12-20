package v1_zone

import (
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/zone_service"
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
	if err := context.GinContext.Bind(&requestQuery); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	zones, err := zone_service.New().WithAuth(auth).List(zone_service.ListOpts{Type: requestQuery.Type})
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	dtoZones := make([]body.ZoneRead, len(zones))
	for i, zone := range zones {
		dtoZones[i] = zone.ToDTO()
	}

	context.Ok(dtoZones)
}

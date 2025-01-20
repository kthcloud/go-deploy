package v2

import (
	"github.com/gin-gonic/gin"
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/dto/v2/query"
	"github.com/kthcloud/go-deploy/pkg/sys"
	"github.com/kthcloud/go-deploy/service"
)

// ListZones
// @Summary List zones
// @Description List zones
// @Tags Zone
// @Produce json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Success 200 {array} body.ZoneRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/zones [get]
func ListZones(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.ZoneList
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, ErrAuthInfoNotAvailable)
		return
	}

	zoneList, err := service.V2(auth).System().ListZones()
	if err != nil {
		context.ServerError(err, ErrInternal)
		return
	}

	dtoZones := make([]body.ZoneRead, len(zoneList))
	for i, zone := range zoneList {
		dtoZones[i] = zone.ToDTO()
	}

	context.Ok(dtoZones)
}

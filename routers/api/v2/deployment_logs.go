package v2

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/dto/v2/uri"
	"github.com/kthcloud/go-deploy/pkg/sys"
	"github.com/kthcloud/go-deploy/service"
	errors2 "github.com/kthcloud/go-deploy/service/errors"
	"github.com/kthcloud/go-deploy/service/v2/deployments"
)

// GetLogs
// @Summary Get logs using Server-Sent Events
// @Description Get logs using Server-Sent Events
// @Tags Deployment
// @Produce json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param deploymentId path string true "Deployment ID"
// @Success 200 {string} string
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/deployments/{deploymentId}/logs [get]
func GetLogs(c *gin.Context) {
	sysContext := sys.NewContext(c)

	var requestURI uri.LogsGet
	if err := sysContext.GinContext.ShouldBindUri(&requestURI); err != nil {
		sysContext.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&sysContext)
	if err != nil {
		sysContext.ServerError(err, ErrAuthInfoNotAvailable)
		return
	}

	type Message struct {
		source    string
		prefix    string
		msg       string
		createdAt time.Time
	}

	ch := make(chan Message)

	handler := func(source, prefix, msg string, createdAt time.Time) {
		ch <- Message{source, prefix, msg, createdAt}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deployV2 := service.V2(auth)

	err = deployV2.Deployments().SetupLogStream(requestURI.DeploymentID, ctx, handler, 25)
	if err != nil {
		if errors.Is(err, errors2.ErrDeploymentNotFound) {
			sysContext.NotFound("Deployment not found")
			return
		}

		sysContext.ServerError(err, ErrInternal)
		return
	}

	c.Stream(func(w io.Writer) bool {
		select {
		case <-ctx.Done():
			return false
		case msg := <-ch:
			if msg.source == deployments.MessageSourceControl {
				return true
			}

			c.SSEvent(msg.source, body.LogMessage{
				Source:    msg.source,
				Prefix:    msg.prefix,
				Line:      msg.msg,
				CreatedAt: msg.createdAt,
			})
			return true
		}
	})
}

// mhuaaaaaaaaaaaah, i love you i love you i love you

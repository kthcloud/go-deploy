package v1_deployment

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"go-deploy/dto/v1/body"
	"go-deploy/dto/v1/uri"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
	errors2 "go-deploy/service/errors"
	"go-deploy/service/v1/deployments"
	"io"
	"time"
)

// GetLogsSSE
// @Summary Get logs SSE
// @Description Get logs using Server-Sent Events
// @Tags Deployment
// @Accept  json
// @Produce  json
// @Param deploymentId path string true "Deployment ID"
// @Success 200 {string} string
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /deployments/{deploymentId}/logs [get]
func GetLogsSSE(c *gin.Context) {
	sysContext := sys.NewContext(c)

	var requestURI uri.LogsGet
	if err := sysContext.GinContext.ShouldBindUri(&requestURI); err != nil {
		sysContext.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&sysContext)
	if err != nil {
		sysContext.ServerError(err, v1.AuthInfoNotAvailableErr)
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

	deployV1 := service.V1(auth)

	err = deployV1.Deployments().SetupLogStream(requestURI.DeploymentID, ctx, handler, 25)
	if err != nil {
		if errors.Is(err, errors2.DeploymentNotFoundErr) {
			sysContext.NotFound("Deployment not found")
		}

		sysContext.ServerError(err, v1.InternalError)
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

//mhuaaaaaaaaaaaah, i love you i love you i love you

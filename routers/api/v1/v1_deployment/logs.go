package v1_deployment

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/uri"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/deployment_service"
	"io"
)

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
		source string
		prefix string
		msg    string
	}

	ch := make(chan Message)

	handler := func(source, prefix, msg string) {
		ch <- Message{source, prefix, msg}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = deployment_service.SetupLogStream(ctx, requestURI.DeploymentID, handler, auth)
	if err != nil {
		if errors.Is(err, deployment_service.DeploymentNotFoundErr) {
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
			if msg.msg == "__control" {
				return true
			}
			//_, err := fmt.Fprintf(w, "data: %s\n\n", fmt.Sprintf("%s: %s", msg.prefix, msg.msg))
			//if err != nil {
			//	utils.PrettyPrintError(fmt.Errorf("error writing message to SSE for deployment %s. details: %w", requestURI.DeploymentID, err))
			//	return false
			//}

			message := fmt.Sprintf("%s: %s", msg.prefix, msg.msg)
			c.SSEvent(msg.source, message)
			return true
		}
	})
}

//mhuaaaaaaaaaaaah, i love you i love you i love you

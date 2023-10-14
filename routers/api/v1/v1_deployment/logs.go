package v1_deployment

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go-deploy/models/dto/uri"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
	"go-deploy/service/deployment_service"
	"go-deploy/utils"
	"go-deploy/utils/requestutils"
	"io"
	"log"
	"net/http"
	"strings"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func GetLogsSSE(c *gin.Context) {
	httpContext := sys.NewContext(c)

	var requestURI uri.LogsGet
	if err := httpContext.GinContext.ShouldBindUri(&requestURI); err != nil {
		httpContext.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&httpContext)
	if err != nil {
		httpContext.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	ch := make(chan string)

	handler := func(msg string) {
		ch <- msg
	}

	ctx, err := deployment_service.SetupLogStream(requestURI.DeploymentID, handler, auth)
	if err != nil {
		httpContext.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if ctx == nil {
		httpContext.ErrorResponse(http.StatusNotFound, status_codes.Error, fmt.Sprintf("deployment %s not found", requestURI.DeploymentID))
		return
	}

	c.Stream(func(w io.Writer) bool {
		select {
		case <-ctx.Done():
			return false
		case msg := <-ch:
			_, err := fmt.Fprintf(w, "data: %s\n\n", msg)
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error writing message to SSE for deployment %s. details: %w", requestURI.DeploymentID, err))
				return false
			}
			return true
		}
	})
}

// GetLogs Websocket
func GetLogs(c *gin.Context) {
	httpContext := sys.NewContext(c)

	var requestURI uri.LogsGet
	if err := httpContext.GinContext.ShouldBindUri(&requestURI); err != nil {
		httpContext.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		httpContext.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	go func() {
		defer func(ws *websocket.Conn) {
			_ = ws.Close()
		}(ws)

		var logContext context.Context
		var auth *service.AuthInfo

		handler := func(msg string) {
			err = ws.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				if strings.Contains(err.Error(), "closed network connection") || strings.Contains(err.Error(), "connection reset by peer") {
					return
				}

				utils.PrettyPrintError(fmt.Errorf("error writing message to websocket for deployment %s. details: %w", requestURI.DeploymentID, err))
				_ = ws.Close()
			}
		}

		for {
			_, data, readMsgErr := ws.ReadMessage()
			var ce *websocket.CloseError
			if errors.As(readMsgErr, &ce) {
				switch ce.Code {
				case websocket.CloseNormalClosure,
					websocket.CloseGoingAway,
					websocket.CloseNoStatusReceived:
					if logContext != nil {
						logContext.Done()
					}
					log.Println("websocket closed for deployment ", requestURI.DeploymentID)
					return
				}
			}

			if readMsgErr != nil {
				utils.PrettyPrintError(fmt.Errorf("error reading message from websocket for deployment %s. details: %w", requestURI.DeploymentID, readMsgErr))
				return
			}

			msg := string(data)

			if strings.HasPrefix(msg, "Bearer ") && auth == nil {
				auth = validateBearerToken(msg)
				if auth != nil {
					logContext, err = deployment_service.SetupLogStream(requestURI.DeploymentID, handler, auth)
					if err != nil {
						httpContext.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
						return
					}

					if logContext == nil {
						utils.PrettyPrintError(fmt.Errorf("deployment not found when trying to setup log stream %s", requestURI.DeploymentID))
						return
					}
				}
			}
		}
	}()
}

func validateBearerToken(bearer string) *service.AuthInfo {
	req, err := http.NewRequest("GET", "http://localhost:8080/v1/authCheck", nil)
	req.Header.Add("Authorization", bearer)
	if err != nil {
		return nil
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var authInfo service.AuthInfo
	err = requestutils.ParseBody(resp.Body, &authInfo)
	if err != nil {
		return nil
	}

	return &authInfo
}

//mhuaaaaaaaaaaaah, i love you i love you i love you

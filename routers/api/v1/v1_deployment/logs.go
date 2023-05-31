package v1_deployment

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go-deploy/models/dto/uri"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/deployment_service"
	"go-deploy/utils/requestutils"
	"log"
	"net/http"
	"strings"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func GetLogs(c *gin.Context) {
	context := app.NewContext(c)

	var requestURI uri.LogsGet
	if err := context.GinContext.BindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	go func() {
		defer func(ws *websocket.Conn) {
			_ = ws.Close()
		}(ws)

		handler := func(msg string) {
			err = ws.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Println("error writing message to websocket for deployment ", requestURI.DeploymentID, ". details:", err)
				_ = ws.Close()
			}
		}

		didAuth := false

		for {
			_, data, readMsgErr := ws.ReadMessage()
			msg := string(data)
			if strings.HasPrefix(msg, "Bearer ") && !didAuth {
				auth := validateBearerToken(msg)
				if auth != nil {
					didAuth = true

					logContext, getLogsErr := deployment_service.GetLogs(auth.UserID, requestURI.DeploymentID, handler, auth.IsAdmin())
					if getLogsErr != nil {
						context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", getLogsErr))
						return
					}

					if logContext == nil {
						context.NotFound()
						return
					}
				}
			}

			if ce, ok := readMsgErr.(*websocket.CloseError); ok {
				switch ce.Code {
				case websocket.CloseNormalClosure,
					websocket.CloseGoingAway,
					websocket.CloseNoStatusReceived:
					log.Println("websocket closed for deployment ", requestURI.DeploymentID)
					return
				}
			}
		}
	}()
}

func validateBearerToken(bearer string) *v1.AuthInfo {
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

	var authInfo v1.AuthInfo
	err = requestutils.ParseBody(resp.Body, &authInfo)
	if err != nil {
		return nil
	}

	return &authInfo
}

package v1_deployment

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/validator"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/deployment_service"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{}

func GetLogs(c *gin.Context) {
	context := app.NewContext(c)

	rules := validator.MapData{
		"deploymentId": []string{"required", "uuid_v4"},
	}

	validationErrors := context.ValidateParams(&rules)

	if len(validationErrors) > 0 {
		context.ResponseValidationError(validationErrors)
		return
	}

	token, err := context.GetKeycloakToken()
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
	}
	userID := token.Sub
	deploymentID := context.GinContext.Param("deploymentId")
	isAdmin := v1.IsAdmin(&context)

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	defer func(ws *websocket.Conn) {
		_ = ws.Close()
	}(ws)

	handler := func(msg string) {
		err = ws.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			fmt.Printf("failed to write websocket message for deployment %s (%s)", deploymentID, ws.RemoteAddr())
			_ = ws.Close()
		}
	}

	logContext, getLogsErr := deployment_service.GetLogs(userID, deploymentID, handler, isAdmin)
	if getLogsErr != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", getLogsErr))
		return
	}

	if logContext == nil {
		context.NotFound()
		return
	}

	for {
		_, _, readMsgErr := ws.ReadMessage()
		if ce, ok := readMsgErr.(*websocket.CloseError); ok {
			switch ce.Code {
			case websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseNoStatusReceived:
				log.Printf("closing websocket connection for deployment %s (%s)\n", deploymentID, ws.RemoteAddr())
				return
			}
		}
	}
}

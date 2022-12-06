package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/validator"
	"go-deploy/service/project_service"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{}

func GetProjectLogs(c *gin.Context) {
	context := app.NewContext(c)

	rules := validator.MapData{
		"userId":    []string{"required", "uuid_v4"},
		"projectId": []string{"required", "uuid_v4"},
	}

	validationErrors := context.ValidateParams(&rules)

	if len(validationErrors) > 0 {
		context.ResponseValidationError(validationErrors)
		return
	}

	userID := context.GinContext.Param("userId")
	projectID := context.GinContext.Param("projectId")
	token, err := context.GetKeycloakToken()
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
	}

	if userID != token.Sub {
		context.Unauthorized()
		return
	}

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
			fmt.Printf("failed to write websocket message for project %s (%s)", projectID, ws.RemoteAddr())
			_ = ws.Close()
		}
	}

	logContext, getLogsErr := project_service.GetLogs(userID, projectID, handler)
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
				log.Printf("closing websocket connection for project %s (%s)\n", projectID, ws.RemoteAddr())
				return
			}
		}
	}
}

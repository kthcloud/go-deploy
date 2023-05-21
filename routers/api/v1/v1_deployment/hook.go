package v1_deployment

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/models/dto/body"
	"go-deploy/models/sys/job"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	"go-deploy/service/deployment_service"
	"go-deploy/service/job_service"
	"go-deploy/utils"
	"go-deploy/utils/requestutils"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func getTokenFromAuthHeader(context app.ClientContext) (string, error) {
	const authHeaderName = "Authorization"

	authHeader := context.GinContext.GetHeader(authHeaderName)
	if len(authHeader) == 0 {
		return "", nil
	}

	headerSplit := strings.Split(authHeader, " ")
	if len(headerSplit) != 2 {
		return "", nil
	}

	if headerSplit[0] != "Basic" {
		return "", nil
	}

	decodedHeader, err := base64.StdEncoding.DecodeString(headerSplit[1])
	if err != nil {
		return "", err
	}

	basicAuthSplit := strings.Split(string(decodedHeader), ":")
	if len(basicAuthSplit) != 2 {
		return "", nil
	}

	return basicAuthSplit[1], nil
}

func HandleHarborHook(c *gin.Context) {
	context := app.ClientContext{GinContext: c}

	token, err := getTokenFromAuthHeader(context)

	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if token == "" {
		context.Unauthorized()
		return
	}

	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	deployment, err := deployment_service.GetByHarborWebhookToken(utils.HashString(token))
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if deployment == nil {
		context.NotFound()
		return
	}

	webhook, err := deployment_service.ParseHarborWebhook(context.GinContext.Request.Body)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if webhook.Type == "PUSH_ARTIFACT" {
		log.Printf("restarting v1_deployment %s due to push\n", deployment.Name)
		err = deployment_service.Restart(deployment.Name)
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
			return
		}
	}

	context.Ok()
}

func HandleGitHubHook(c *gin.Context) {
	context := app.ClientContext{GinContext: c}

	event := context.GinContext.GetHeader("x-github-event")
	if len(event) == 0 {
		context.ErrorResponse(http.StatusBadRequest, status_codes.Error, "missing x-github-event header")
		return
	}

	if event == "ping" {
		context.Ok()
		return
	}

	if event != "push" {
		context.ErrorResponse(http.StatusBadRequest, status_codes.Error, "unsupported event type")
		return
	}

	hookIdStr := context.GinContext.GetHeader("X-Github-Hook-Id")
	if len(hookIdStr) == 0 {
		context.ErrorResponse(http.StatusBadRequest, status_codes.Error, "missing x-github-delivery header")
		return
	}

	hookID, err := strconv.ParseInt(hookIdStr, 10, 64)
	if err != nil {
		context.ErrorResponse(http.StatusBadRequest, status_codes.Error, "invalid x-github-delivery header")
		return
	}

	signature := context.GinContext.GetHeader("x-hub-signature-256")
	if len(signature) == 0 {
		context.ErrorResponse(http.StatusBadRequest, status_codes.Error, "missing x-hub-signature-256 header")
		return
	}

	requestBodyRaw, err := requestutils.ReadBody(context.GinContext.Request.Body)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	deployment, err := deployment_service.GetByGitHubWebhookID(hookID)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if !checkSignature(signature, requestBodyRaw, deployment.Subsystems.GitHub.Webhook.Secret) {
		context.ErrorResponse(http.StatusBadRequest, status_codes.Error, "invalid signature")
		return
	}

	var requestBodyParsed body.GithubWebhookPayloadPush
	err = json.Unmarshal(requestBodyRaw, &requestBodyParsed)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	pushedBranch := strings.Split(requestBodyParsed.Ref, "/")[2]
	if pushedBranch != requestBodyParsed.Repository.DefaultBranch {
		context.Ok()
		return
	}

	jobID := uuid.NewString()
	err = job_service.Create(jobID, deployment.OwnerID, job.TypeBuildDeployment, map[string]interface{}{
		"id": deployment.ID,
		"build": body.DeploymentBuild{
			Tag:       "latest",
			Branch:    pushedBranch,
			ImportURL: requestBodyParsed.Repository.CloneURL,
		},
	})

	context.Ok()
}

func checkSignature(signature string, payload []byte, secret string) bool {
	const signaturePrefix = "sha256="
	const prefixLength = len(signaturePrefix)
	const signatureLength = prefixLength + (sha256.Size * 2)

	if len(signature) != signatureLength || !strings.HasPrefix(signature, signaturePrefix) {
		return false
	}

	actual := make([]byte, sha256.Size)
	_, _ = hex.Decode(actual, []byte(signature[prefixLength:]))

	byteStringSecret := []byte(secret)

	expected := getSignature(byteStringSecret, payload)

	return hmac.Equal(expected, actual)
}

func getSignature(secret, body []byte) []byte {
	computed := hmac.New(sha256.New, secret)
	computed.Write(body)
	return []byte(computed.Sum(nil))
}

package v1_deployment

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/models/dto/body"
	"go-deploy/models/sys/job"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/deployment_service"
	"go-deploy/service/job_service"
	"go-deploy/utils/requestutils"
	"strconv"
	"strings"
)

func getTokenFromAuthHeader(context sys.ClientContext) (string, error) {
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
	context := sys.NewContext(c)

	token, err := getTokenFromAuthHeader(context)

	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if token == "" {
		context.Unauthorized("Missing token")
		return
	}

	if !deployment_service.ValidateHarborToken(token) {
		context.Unauthorized("Invalid token")
		return
	}

	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	var webhook body.HarborWebhook
	err = context.GinContext.ShouldBindJSON(&webhook)
	if err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	deployment, err := deployment_service.GetByHarborWebhook(&webhook)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if deployment == nil {
		context.NotFound("Deployment not found")
		return
	}

	if webhook.Type == "PUSH_ARTIFACT" {
		err = deployment_service.Restart(deployment.ID)
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}
	}

	context.OkNoContent()
}

func HandleGitHubHook(c *gin.Context) {
	context := sys.NewContext(c)

	event := context.GinContext.GetHeader("x-github-event")
	if len(event) == 0 {
		context.UserError("Missing x-github-event header")
		return
	}

	if event == "ping" {
		context.OkNoContent()
		return
	}

	if event != "push" {
		context.UserError("Unsupported event type")
		return
	}

	hookIdStr := context.GinContext.GetHeader("X-Github-Hook-Id")
	if len(hookIdStr) == 0 {
		context.UserError("Missing X-GitHub-Hook-Id header")
		return
	}

	hookID, err := strconv.ParseInt(hookIdStr, 10, 64)
	if err != nil {
		context.UserError("Invalid X-GitHub-Hook-Id header")
		return
	}

	if hookID == 0 {
		context.UserError("Invalid X-GitHub-Hook-Id header")
		return
	}

	signature := context.GinContext.GetHeader("x-hub-signature-256")
	if len(signature) == 0 {
		context.UserError("Missing x-hub-signature-256 header")
		return
	}

	requestBodyRaw, err := requestutils.ReadBody(context.GinContext.Request.Body)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	var requestBodyParsed body.GithubWebhookPayloadPush
	if err = context.GinContext.ShouldBindJSON(&requestBodyParsed); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	deployments, err := deployment_service.ListByGitHubWebhookID(hookID)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if len(deployments) == 0 {
		context.NotFound("No deployments found for hook ID")
		return
	}

	var ids []string
	for _, deployment := range deployments {
		if !checkSignature(signature, requestBodyRaw, deployment.Subsystems.GitHub.Webhook.Secret) {
			context.Forbidden("Invalid signature")
			return
		}

		ids = append(ids, deployment.ID)
	}

	refSplit := strings.Split(requestBodyParsed.Ref, "/")
	if len(refSplit) != 3 {
		context.UserError("Invalid ref field")
		return
	}

	pushedBranch := refSplit[2]
	if pushedBranch != requestBodyParsed.Repository.DefaultBranch {
		// We only care about the default branch, so this is not an error
		context.OkNoContent()
		return
	}

	if len(ids) > 0 {
		jobID := uuid.NewString()
		err = job_service.Create(jobID, "system", job.TypeBuildDeployments, map[string]interface{}{
			"ids": ids,
			"build": body.DeploymentBuild{
				Name:      requestBodyParsed.Repository.Name,
				Tag:       "latest",
				Branch:    pushedBranch,
				ImportURL: requestBodyParsed.Repository.CloneURL,
			},
		})
	}

	context.OkNoContent()
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

package models

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	modelv2 "github.com/mittwald/goharbor-client/v5/apiv2/model"
	"golang.org/x/crypto/bcrypt"
	"math/big"
	"strings"
)

type WebhookPublic struct {
	ID          int    `json:"id" bson:"id"`
	Name        string `json:"name" bson:"name"`
	ProjectID   int    `json:"projectId" bson:"projectId"`
	ProjectName string `json:"projectName" bson:"projectName"`
	Target      string `json:"target" bson:"target"`
	Token       string `json:"token" bson:"token"`
}

func (w *WebhookPublic) Created() bool {
	return w.ID != 0
}

func CreateWebhookParamsFromPublic(public *WebhookPublic) *modelv2.WebhookPolicy {
	return &modelv2.WebhookPolicy{
		Enabled:    true,
		EventTypes: getWebhookEventTypes(),
		Name:       public.Name,
		Targets: []*modelv2.WebhookTargetObject{
			{
				Address:        public.Target,
				AuthHeader:     createWebhookAuthHeader(public.Token),
				SkipCertVerify: false,
				Type:           "http",
			},
		},
	}
}

func CreateWebhookPublicFromGet(webhookPolicy *modelv2.WebhookPolicy, project *modelv2.Project) *WebhookPublic {
	token := getTokenFromAuthHeader(webhookPolicy.Targets[0].AuthHeader)

	return &WebhookPublic{
		ID:          int(webhookPolicy.ID),
		Name:        webhookPolicy.Name,
		ProjectID:   int(project.ProjectID),
		ProjectName: project.Name,
		Target:      webhookPolicy.Targets[0].Address,
		Token:       token,
	}
}

func getWebhookEventTypes() []string {
	return []string{
		"PUSH_ARTIFACT",
	}
}

func createWebhookAuthHeader(secret string) string {
	fullSecret := fmt.Sprintf("cloud:%s", secret)
	encoded := base64.StdEncoding.EncodeToString([]byte(fullSecret))
	return fmt.Sprintf("Basic %s", encoded)
}

func getTokenFromAuthHeader(authHeader string) string {
	if len(authHeader) == 0 {
		return ""
	}

	headerSplit := strings.Split(authHeader, " ")
	if len(headerSplit) != 2 {
		return ""
	}

	if headerSplit[0] != "Basic" {
		return ""
	}

	decodedHeader, err := base64.StdEncoding.DecodeString(headerSplit[1])
	if err != nil {
		return ""
	}

	basicAuthSplit := strings.Split(string(decodedHeader), ":")
	if len(basicAuthSplit) != 2 {
		return ""
	}

	return basicAuthSplit[1]
}

func generateWebhookToken(secret string) (string, error) {
	salt, err := generateRandomString(10)
	if err != nil {
		return "", err
	}

	saltedSecret := fmt.Sprintf("%s%s", secret, salt)
	hash, err := bcrypt.GenerateFromPassword([]byte(saltedSecret), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	hasher := md5.New()
	hasher.Write(hash)
	encoded := hex.EncodeToString(hasher.Sum(nil))
	return encoded, nil
}

func generateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}

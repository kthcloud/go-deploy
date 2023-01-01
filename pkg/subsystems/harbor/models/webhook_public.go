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
)

type WebhookPublic struct {
	ID            int    `json:"id" bson:"id"`
	Name          string `json:"name" bson:"name"`
	ProjectID     int    `json:"projectId" bson:"projectId"`
	ProjectName   string `json:"projectName" bson:"projectName"`
	Target string `json:"target" bson:"target"`
	Token  string `json:"token" bson:"token"`
}

func CreateWebhookParamsFromPublic(public *WebhookPublic) *modelv2.WebhookPolicy {
	webhookToken, err := generateWebhookToken(public.Token)
	if err != nil {
		webhookToken = ""
	}

	return &modelv2.WebhookPolicy{
		Enabled:    true,
		EventTypes: getWebhookEventTypes(),
		Name:       public.Name,
		Targets: []*modelv2.WebhookTargetObject{
			{
				Address:        public.Target,
				AuthHeader:     createWebhookAuthHeader(webhookToken),
				SkipCertVerify: false,
				Type:           "http",
			},
		},
	}
}

func CreateWebhookPublicFromGet(webhookPolicy *modelv2.WebhookPolicy, project *modelv2.Project) *WebhookPublic {
	return &WebhookPublic{
		ID:          int(webhookPolicy.ID),
		Name:        webhookPolicy.Name,
		ProjectID:   int(project.ProjectID),
		ProjectName: project.Name,
		Target:      webhookPolicy.Targets[0].Address,
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

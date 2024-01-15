package models

import (
	"encoding/base64"
	"fmt"
	modelv2 "github.com/mittwald/goharbor-client/v5/apiv2/model"
	"strings"
	"time"
)

type WebhookPublic struct {
	ID        int       `bson:"id"`
	Name      string    `bson:"name"`
	Target    string    `bson:"target"`
	Token     string    `bson:"token"`
	CreatedAt time.Time `bson:"createdAt"`
}

func (w *WebhookPublic) Created() bool {
	return w.ID != 0
}

func (w *WebhookPublic) IsPlaceholder() bool {
	return false
}

// CreateWebhookParamsFromPublic creates a body used for creating a webhook in the Harbor API.
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

// CreateWebhookUpdateParamsFromPublic creates a body used for updating a webhook in the Harbor API.
func CreateWebhookUpdateParamsFromPublic(public *WebhookPublic, current *modelv2.WebhookPolicy) *modelv2.WebhookPolicy {
	update := *current
	update.Enabled = true
	update.Name = public.Name
	update.EventTypes = getWebhookEventTypes()
	update.Targets = []*modelv2.WebhookTargetObject{
		{
			Address:        public.Target,
			AuthHeader:     createWebhookAuthHeader(public.Token),
			SkipCertVerify: false,
			Type:           "http",
		},
	}

	return &update
}

// CreateWebhookPublicFromGet converts a modelv2.WebhookPolicy to a WebhookPublic.
func CreateWebhookPublicFromGet(webhookPolicy *modelv2.WebhookPolicy, project *modelv2.Project) *WebhookPublic {
	token := getTokenFromAuthHeader(webhookPolicy.Targets[0].AuthHeader)

	return &WebhookPublic{
		ID:        int(webhookPolicy.ID),
		Name:      webhookPolicy.Name,
		Target:    webhookPolicy.Targets[0].Address,
		Token:     token,
		CreatedAt: time.Time(webhookPolicy.CreationTime),
	}
}

// getWebhookEventTypes returns the event types that should be listened to by the webhook.
func getWebhookEventTypes() []string {
	return []string{
		"PUSH_ARTIFACT",
	}
}

// createWebhookAuthHeader creates an auth header for the webhook.
func createWebhookAuthHeader(secret string) string {
	fullSecret := fmt.Sprintf("cloud:%s", secret)
	encoded := base64.StdEncoding.EncodeToString([]byte(fullSecret))
	return fmt.Sprintf("Basic %s", encoded)
}

// getTokenFromAuthHeader extracts the token from an auth header.
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

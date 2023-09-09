package models

import (
	"github.com/google/go-github/github"
	"time"
)

type WebhookPublic struct {
	ID           int64     `bson:"id"`
	RepositoryID int64     `bson:"repositoryId"`
	Events       []string  `bson:"events"`
	Active       bool      `bson:"active"`
	ContentType  string    `bson:"contentType"`
	WebhookURL   string    `bson:"webhook"`
	Secret       string    `bson:"secret"`
	CreatedAt    time.Time `bson:"createdAt"`
}

func (w *WebhookPublic) Created() bool {
	return w.ID != 0
}

func CreateWebhookPublicFromGet(webhook *github.Hook, repositoryID int64) *WebhookPublic {
	contentType := ""
	if webhook.Config["content_type"] != nil {
		contentType = webhook.Config["content_type"].(string)
	}

	webhookURL := ""
	if webhook.Config["url"] != nil {
		webhookURL = webhook.Config["url"].(string)
	}

	secret := ""
	if webhook.Config["token"] != nil {
		secret = webhook.Config["token"].(string)
	}

	return &WebhookPublic{
		ID:           webhook.GetID(),
		RepositoryID: repositoryID,
		Events:       webhook.Events,
		Active:       webhook.GetActive(),
		ContentType:  contentType,
		WebhookURL:   webhookURL,
		Secret:       secret,
		CreatedAt:    webhook.GetCreatedAt(),
	}
}

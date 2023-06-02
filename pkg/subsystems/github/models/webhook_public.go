package models

import "github.com/google/go-github/github"

type WebhookPublic struct {
	ID           int64    `bson:"id"`
	RepositoryID int64    `bson:"repositoryId"`
	Events       []string `bson:"events"`
	Active       bool     `bson:"active"`
	ContentType  string   `bson:"contentType"`
	WebhookURL   string   `bson:"webhook"`
	Secret       string   `bson:"secret"`
}

func (w *WebhookPublic) Created() bool {
	return w.ID != 0
}

func CreateWebhookPublicFromGet(webhook *github.Hook, repositoryID int64) *WebhookPublic {
	return &WebhookPublic{
		ID:           *webhook.ID,
		RepositoryID: repositoryID,
		Events:       webhook.Events,
		Active:       *webhook.Active,
		ContentType:  webhook.Config["content_type"].(string),
		WebhookURL:   webhook.Config["url"].(string),
		Secret:       webhook.Config["token"].(string),
	}
}

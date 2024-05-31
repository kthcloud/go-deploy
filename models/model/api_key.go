package model

import (
	"go-deploy/dto/v2/body"
	"time"
)

type ApiKey struct {
	Name      string    `bson:"name"`
	Key       string    `bson:"key"`
	CreatedAt time.Time `bson:"createdAt"`
	ExpiresAt time.Time `bson:"expiresAt"`
}

type ApiKeyCreateParams struct {
	Name      string    `json:"name"`
	Key       string    `json:"key"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func (apiKey *ApiKey) ToDTO() body.ApiKeyCreated {
	return body.ApiKeyCreated{
		Name:      apiKey.Name,
		Key:       apiKey.Key,
		CreatedAt: apiKey.CreatedAt,
		ExpiresAt: apiKey.ExpiresAt,
	}
}

func (apiKey ApiKeyCreateParams) FromDTO(dto *body.ApiKeyCreate, key string) *ApiKeyCreateParams {
	return &ApiKeyCreateParams{
		Name:      dto.Name,
		Key:       key,
		ExpiresAt: dto.ExpiresAt,
	}
}

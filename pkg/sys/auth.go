package sys

import (
	"encoding/json"
	"fmt"
	"go-deploy/pkg/auth"
	"go-deploy/pkg/config"
)

// HasApiKey checks if the request has an api key.
func (context *ClientContext) HasApiKey() bool {
	return context.GinContext.GetHeader("X-API-KEY") != ""
}

// GetApiKey gets the api key from the request.
func (context *ClientContext) GetApiKey() (string, error) {
	apiKey := context.GinContext.GetHeader("X-API-KEY")
	if apiKey == "" {
		return "", fmt.Errorf("failed to find api key in request")
	}

	return apiKey, nil
}

// HasBearerToken checks if the request has a bearer token.
func (context *ClientContext) HasBearerToken() bool {
	return context.GinContext.GetHeader("Authorization") != ""
}

// HasKeycloakToken checks if the request has a keycloak token.
func (context *ClientContext) HasKeycloakToken() bool {
	_, exists := context.GinContext.Get("keycloakToken")
	return exists
}

// GetKeycloakToken gets the keycloak token from the request.
func (context *ClientContext) GetKeycloakToken() (*auth.KeycloakToken, error) {
	tokenRaw, exists := context.GinContext.Get("keycloakToken")
	if !exists {
		return nil, fmt.Errorf("failed to find token in request")
	}

	bytes, err := json.Marshal(tokenRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token in request")
	}

	keycloakToken := auth.KeycloakToken{}
	err = json.Unmarshal(bytes, &keycloakToken)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token in request")
	}

	return &keycloakToken, nil
}

// GetKeyCloakConfig gets the keycloak config.
func GetKeyCloakConfig() auth.KeycloakConfig {
	var fullCertPath = fmt.Sprintf("realms/%s/protocol/openid-connect/certs", config.Config.Keycloak.Realm)

	return auth.KeycloakConfig{
		Url:           config.Config.Keycloak.Url,
		Realm:         config.Config.Keycloak.Realm,
		FullCertsPath: &fullCertPath}
}

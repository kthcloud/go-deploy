package app

import (
	"encoding/json"
	"fmt"
	"go-deploy/pkg/auth"
	"go-deploy/pkg/conf"
)

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

func GetKeyCloakConfig() auth.KeycloakConfig {
	var fullCertPath = fmt.Sprintf("realms/%s/protocol/openid-connect/certs", conf.Env.Keycloak.Realm)

	return auth.KeycloakConfig{
		Url:           conf.Env.Keycloak.Url,
		Realm:         conf.Env.Keycloak.Realm,
		FullCertsPath: &fullCertPath}
}

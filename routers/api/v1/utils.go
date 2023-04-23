package v1

import (
	"go-deploy/pkg/app"
	"go-deploy/pkg/conf"
)

func IsAdmin(context *app.ClientContext) bool {
	token, err := context.GetKeycloakToken()
	if err != nil {
		return false
	}

	for _, group := range token.Groups {
		if group == conf.Env.Keycloak.AdminGroup {
			return true
		}
	}

	return false
}

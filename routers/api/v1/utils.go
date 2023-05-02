package v1

import (
	"go-deploy/pkg/app"
	"go-deploy/pkg/conf"
)

func IsAdmin(context *app.ClientContext) bool {
	return InGroup(context, conf.Env.Keycloak.AdminGroup)
}

func IsPowerUser(context *app.ClientContext) bool {
	return InGroup(context, conf.Env.Keycloak.PowerUserGroup)
}

func InGroup(context *app.ClientContext, group string) bool {
	token, err := context.GetKeycloakToken()
	if err != nil {
		return false
	}

	for _, userGroup := range token.Groups {
		if userGroup == group {
			return true
		}
	}

	return false
}

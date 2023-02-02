package user_info_service

import (
	"go-deploy/models/user_info"
	"go-deploy/pkg/auth"
)

func GetByToken(token *auth.KeycloakToken) (*user_info.UserInfo, error) {
	err := user_info.CreateEmpty(token)
	if err != nil {
		return nil, err
	}

	return user_info.GetBySub(token.Sub)
}

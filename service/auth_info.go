package service

import (
	roleModel "go-deploy/models/config/role"
	"go-deploy/pkg/auth"
	"go-deploy/pkg/config"
)

type AuthInfo struct {
	UserID   string              `json:"userId"`
	JwtToken *auth.KeycloakToken `json:"jwtToken"`
	Roles    []roleModel.Role    `json:"roles"`
	IsAdmin  bool                `json:"isAdmin"`
}

func CreateAuthInfo(userID string, JwtToken *auth.KeycloakToken, iamGroups []string) *AuthInfo {
	roles := config.Config.GetRolesByIamGroups(iamGroups)

	isAdmin := false
	for _, iamGroup := range iamGroups {
		if iamGroup == config.Config.Keycloak.AdminGroup {
			isAdmin = true
		}
	}

	return &AuthInfo{
		UserID:   userID,
		JwtToken: JwtToken,
		Roles:    roles,
		IsAdmin:  isAdmin,
	}
}

func (authInfo *AuthInfo) GetEffectiveRole() *roleModel.Role {
	// roles are assumed to be given in order of priority, weak -> strong
	// so, we can safely return the last one
	if len(authInfo.Roles) == 0 {
		return config.Config.GetRole("default")
	}

	return &authInfo.Roles[len(authInfo.Roles)-1]
}

func (authInfo *AuthInfo) GetUsername() string {
	return authInfo.JwtToken.PreferredUsername
}

func (authInfo *AuthInfo) GetFirstName() string {
	return authInfo.JwtToken.GivenName
}

func (authInfo *AuthInfo) GetLastName() string {
	return authInfo.JwtToken.FamilyName
}

func (authInfo *AuthInfo) GetEmail() string {
	return authInfo.JwtToken.Email
}

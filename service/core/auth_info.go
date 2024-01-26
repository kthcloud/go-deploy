package core

import (
	roleModels "go-deploy/models/sys/role"
	"go-deploy/pkg/auth"
	"go-deploy/pkg/config"
)

// AuthInfo is used to pass auth info to services
// It is used to perform authorization checks.
type AuthInfo struct {
	UserID   string              `json:"userId"`
	JwtToken *auth.KeycloakToken `json:"jwtToken"`
	Roles    []roleModels.Role   `json:"roles"`
	IsAdmin  bool                `json:"isAdmin"`
}

// CreateAuthInfo creates an AuthInfo object
func CreateAuthInfo(userID string, jwtToken *auth.KeycloakToken, iamGroups []string) *AuthInfo {
	roles := config.Config.GetRolesByIamGroups(iamGroups)

	isAdmin := false
	for _, iamGroup := range iamGroups {
		if iamGroup == config.Config.Keycloak.AdminGroup {
			isAdmin = true
		}
	}

	return &AuthInfo{
		UserID:   userID,
		JwtToken: jwtToken,
		Roles:    roles,
		IsAdmin:  isAdmin,
	}
}

// GetEffectiveRole gets the effective role of the user
// This is effectively the strongest role the user has
func (authInfo *AuthInfo) GetEffectiveRole() *roleModels.Role {
	// roles are assumed to be given in order of priority, weak -> strong
	// so, we can safely return the last one
	if len(authInfo.Roles) == 0 {
		defaultRole := config.Config.GetRole("default")
		if defaultRole == nil {
			panic("default role not found")
		}

		return defaultRole
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

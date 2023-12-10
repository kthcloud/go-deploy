package service

import (
	"fmt"
	roleModel "go-deploy/models/sys/role"
	userModel "go-deploy/models/sys/user"
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

func CreateAuthInfoFromDB(userID string) (*AuthInfo, error) {
	user, err := userModel.New().GetByID(userID)
	if err != nil {
		return nil, err
	}

	role := config.Config.GetRole(user.EffectiveRole.Name)
	if role == nil {
		return nil, fmt.Errorf("failed to get role from db effective role %s", user.EffectiveRole.Name)
	}

	return &AuthInfo{
		UserID:   user.ID,
		JwtToken: nil,
		Roles:    []roleModel.Role{*role},
		IsAdmin:  user.IsAdmin,
	}, nil
}

func (authInfo *AuthInfo) GetEffectiveRole() *roleModel.Role {
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

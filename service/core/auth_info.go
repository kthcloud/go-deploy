package core

import (
	"go-deploy/models/model"
	"go-deploy/pkg/config"
)

// AuthInfo is used to pass auth info to services
// It is used to perform authorization checks.
type AuthInfo struct {
	User *model.User `json:"user"`
}

// CreateAuthInfo creates an AuthInfo object
func CreateAuthInfo(user *model.User) *AuthInfo {
	return &AuthInfo{
		User: user,
	}
}

// GetEffectiveRole gets the effective role of the user
// This is effectively the strongest role the user has
func (authInfo *AuthInfo) GetEffectiveRole() *model.Role {
	effectiveRole := config.Config.GetRole(authInfo.User.EffectiveRole.Name)
	if effectiveRole != nil {
		return effectiveRole
	}

	defaultRole := config.Config.GetRole("default")
	if defaultRole == nil {
		panic("default role not found")
	}

	return defaultRole
}

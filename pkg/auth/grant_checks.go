package auth

import (
	"slices"

	"github.com/gin-gonic/gin"
)

type AccessCheckFunction func(tc *TokenContainer, ctx *gin.Context) bool

type AccessTuple struct {
	Service string
	Role    string
	Uid     string
}

func GroupCheck(at []AccessTuple) func(tc *TokenContainer, ctx *gin.Context) bool {
	ats := at
	return func(tc *TokenContainer, ctx *gin.Context) bool {
		addTokenToContext(tc, ctx)
		for idx := range ats {
			at := ats[idx]
			if tc.KeyCloakToken.ResourceAccess != nil {
				serviceRoles := tc.KeyCloakToken.ResourceAccess[at.Service]
				if slices.Contains(serviceRoles.Roles, at.Role) {
					return true
				}
			}
		}
		return false
	}
}

func RealmCheck(allowedRoles []string) func(tc *TokenContainer, ctx *gin.Context) bool {

	return func(tc *TokenContainer, ctx *gin.Context) bool {
		addTokenToContext(tc, ctx)
		for _, allowedRole := range allowedRoles {
			if slices.Contains(tc.KeyCloakToken.RealmAccess.Roles, allowedRole) {
				return true
			}
		}
		return false
	}
}

func addTokenToContext(tc *TokenContainer, ctx *gin.Context) {
	ctx.Set("keycloakToken", *tc.KeyCloakToken)
	ctx.Set("uid", tc.KeyCloakToken.PreferredUsername)
}

func UidCheck(at []AccessTuple) func(tc *TokenContainer, ctx *gin.Context) bool {
	ats := at
	return func(tc *TokenContainer, ctx *gin.Context) bool {
		addTokenToContext(tc, ctx)
		uid := tc.KeyCloakToken.PreferredUsername
		for idx := range ats {
			at := ats[idx]
			if at.Uid == uid {
				return true
			}
		}
		return false
	}
}

func Check() func(tc *TokenContainer, ctx *gin.Context) bool {
	return func(tc *TokenContainer, ctx *gin.Context) bool {
		addTokenToContext(tc, ctx)
		return true
	}
}

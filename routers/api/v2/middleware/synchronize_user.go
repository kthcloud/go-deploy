package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/sys"
	v2 "github.com/kthcloud/go-deploy/routers/api/v2"
	"github.com/kthcloud/go-deploy/service"
)

// SetupAuthUser is a middleware that sets up the authenticated user in the context.
// This is necessary for the authorization checks to work.
// It also synchronizes the users in the database with the users in the keycloak server.
func SetupAuthUser(c *gin.Context) {
	context := sys.NewContext(c)

	var user *model.User
	switch {
	default:
		context.Unauthorized("No authentication method provided")
		c.Abort()
		return

	case context.HasApiKey():
		apiKey, err := context.GetApiKey()
		if err != nil {
			context.ServerError(err, v2.ErrAuthInfoSetupFailed)
			c.Abort()
			return
		}

		user, err = service.V2().Users().GetByApiKey(apiKey)
		if err != nil {
			context.ServerError(err, v2.ErrAuthInfoSetupFailed)
			c.Abort()
			return
		}

		if user == nil {
			context.Unauthorized("Invalid or expired API key")
			c.Abort()
			return
		}
	case context.HasKeycloakToken():
		jwtToken, err := context.GetKeycloakToken()
		if err != nil {
			context.ServerError(err, v2.ErrInternal)
			c.Abort()
			return
		}

		isAdmin := false
		for _, group := range jwtToken.Groups {
			if group == config.Config.Keycloak.AdminGroup {
				isAdmin = true
				break
			}
		}

		authParams := &model.AuthParams{
			UserID:    jwtToken.Sub,
			Username:  jwtToken.PreferredUsername,
			FirstName: jwtToken.Name,
			LastName:  jwtToken.FamilyName,
			Email:     jwtToken.Email,
			IsAdmin:   isAdmin,
			Roles:     config.Config.GetRolesByIamGroups(jwtToken.Groups),
		}

		user, err = service.V2().Users().Synchronize(authParams)
		if err != nil {
			context.ServerError(err, v2.ErrAuthInfoSetupFailed)
			c.Abort()
			return
		}

		if user == nil {
			context.ServerError(fmt.Errorf("failed to synchronize auth user"), v2.ErrAuthInfoSetupFailed)
			c.Abort()
			return
		}
	}

	if user == nil {
		context.ServerError(fmt.Errorf("failed to synchronize auth user"), v2.ErrAuthInfoSetupFailed)
		c.Abort()
		return
	}

	c.Set("authUser", user)
	c.Next()
}

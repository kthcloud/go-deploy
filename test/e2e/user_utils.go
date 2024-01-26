package e2e

import (
	"go-deploy/models/dto/v1/body"
	"testing"
)

func GetUser(t *testing.T, id string) body.UserRead {
	resp := DoGetRequest(t, "/users/"+id)
	return Parse[body.UserRead](t, resp)
}

func ListUsers(t *testing.T, query string) []body.UserRead {
	resp := DoGetRequest(t, "/users"+query)
	return Parse[[]body.UserRead](t, resp)
}

func ListUsersDiscovery(t *testing.T, query string) []body.UserReadDiscovery {
	if query == "" {
		query = "?"
	}

	resp := DoGetRequest(t, "/users"+query+"&discovery=true")
	return Parse[[]body.UserReadDiscovery](t, resp)
}

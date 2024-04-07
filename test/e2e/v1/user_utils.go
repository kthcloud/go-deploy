package v1

import (
	"go-deploy/dto/v1/body"
	"go-deploy/test/e2e"
	"testing"
)

func GetUser(t *testing.T, id string) body.UserRead {
	resp := e2e.DoGetRequest(t, "/users/"+id)
	return e2e.Parse[body.UserRead](t, resp)
}

func ListUsers(t *testing.T, query string) []body.UserRead {
	resp := e2e.DoGetRequest(t, "/users"+query)
	return e2e.Parse[[]body.UserRead](t, resp)
}

func ListUsersDiscovery(t *testing.T, query string) []body.UserReadDiscovery {
	if query == "" {
		query = "?"
	}

	resp := e2e.DoGetRequest(t, "/users"+query+"&discovery=true")
	return e2e.Parse[[]body.UserReadDiscovery](t, resp)
}

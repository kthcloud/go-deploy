package v1

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/dto/v1/body"
	"go-deploy/test"
	"go-deploy/test/e2e"
	"net/http"
	"testing"
)

const (
	TeamPath  = "/v1/teams/"
	TeamsPath = "/v1/teams"
)

func GetTeam(t *testing.T, id string, userID ...string) body.TeamRead {
	resp := e2e.DoGetRequest(t, TeamPath+id, userID...)
	return e2e.MustParse[body.TeamRead](t, resp)
}

func ListTeams(t *testing.T, query string, userID ...string) []body.TeamRead {
	resp := e2e.DoGetRequest(t, TeamsPath+query, userID...)
	return e2e.MustParse[[]body.TeamRead](t, resp)
}

func UpdateTeam(t *testing.T, id string, teamUpdate body.TeamUpdate, userID ...string) body.TeamRead {
	resp := e2e.DoPostRequest(t, TeamPath+id, teamUpdate, userID...)
	var teamRead body.TeamRead
	err := e2e.ReadResponseBody(t, resp, &teamRead)
	assert.NoError(t, err, "team was not updated")

	if teamUpdate.Name != nil {
		assert.Equal(t, *teamUpdate.Name, teamRead.Name, "invalid team name")
	}

	if teamUpdate.Description != nil {
		assert.Equal(t, *teamUpdate.Description, teamRead.Description, "invalid team description")
	}

	if teamUpdate.Resources != nil {
		var requested []string
		var result []string

		for _, resource := range *teamUpdate.Resources {
			requested = append(requested, resource)
		}

		for _, resource := range teamRead.Resources {
			result = append(result, resource.ID)
		}

		test.EqualOrEmpty(t, requested, result, "invalid team resources")
	}

	if teamUpdate.Members != nil {
		var requested []string
		var result []string

		if len(userID) > 0 {
			requested = append(requested, userID[0])
		} else {
			requested = append(requested, e2e.AdminUserID)
		}

		for _, member := range *teamUpdate.Members {
			if member.ID == e2e.AdminUserID {
				continue
			}

			requested = append(requested, member.ID)
		}

		for _, member := range teamRead.Members {
			result = append(result, member.ID)
		}

		test.EqualOrEmpty(t, requested, result, "invalid team members")
	}

	return teamRead
}

func DeleteTeam(t *testing.T, id string, userID ...string) {
	resp := e2e.DoDeleteRequest(t, TeamPath+id, userID...)
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		t.Errorf("team was not deleted")
	}
}

func JoinTeam(t *testing.T, id string, invitationCode string, userID ...string) {
	teamJoin := body.TeamJoin{
		InvitationCode: invitationCode,
	}

	resp := e2e.DoPostRequest(t, TeamPath+id, teamJoin, userID...)
	_ = e2e.MustParse[struct{}](t, resp)
}

func WithTeam(t *testing.T, teamCreate body.TeamCreate, userID ...string) body.TeamRead {
	var requestedMembers []string
	if len(userID) > 0 {
		requestedMembers = append(requestedMembers, userID[0])
	} else {
		requestedMembers = append(requestedMembers, e2e.PowerUserID)
	}

	for _, member := range teamCreate.Members {
		if member.ID == e2e.PowerUserID {
			continue
		}

		requestedMembers = append(requestedMembers, member.ID)
	}
	requestedResources := teamCreate.Resources

	var createdMembers []string
	var createdResources []string

	resp := e2e.DoPostRequest(t, TeamsPath, teamCreate, userID...)
	teamRead := e2e.MustParse[body.TeamRead](t, resp)

	assert.Equal(t, teamCreate.Name, teamRead.Name, "invalid team name")
	assert.Equal(t, teamCreate.Description, teamRead.Description, "invalid team description")

	for _, member := range teamRead.Members {
		createdMembers = append(createdMembers, member.ID)
	}

	for _, resource := range teamRead.Resources {
		createdResources = append(createdResources, resource.ID)
	}

	test.EqualOrEmpty(t, requestedMembers, createdMembers, "invalid team members")
	test.EqualOrEmpty(t, requestedResources, createdResources, "invalid team resources")

	t.Cleanup(func() {
		e2e.DoDeleteRequest(t, TeamPath+teamRead.ID)
	})

	return teamRead
}

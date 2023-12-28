package e2e

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/body"
	"net/http"
	"testing"
)

func GetTeam(t *testing.T, id string, userID ...string) body.TeamRead {
	resp := DoGetRequest(t, "/teams/"+id, userID...)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "team was not fetched")

	var teamRead body.TeamRead
	err := ReadResponseBody(t, resp, &teamRead)
	assert.NoError(t, err, "team was not fetched")

	return teamRead
}

func ListTeams(t *testing.T, query string, userID ...string) []body.TeamRead {
	resp := DoGetRequest(t, "/teams"+query, userID...)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "teams were not fetched")

	var teams []body.TeamRead
	err := ReadResponseBody(t, resp, &teams)
	assert.NoError(t, err, "teams were not fetched")

	return teams
}

func UpdateTeam(t *testing.T, id string, teamUpdate body.TeamUpdate, userID ...string) body.TeamRead {
	resp := DoPostRequest(t, "/teams/"+id, teamUpdate, userID...)
	var teamRead body.TeamRead
	err := ReadResponseBody(t, resp, &teamRead)
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

		EqualOrEmpty(t, requested, result, "invalid team resources")
	}

	if teamUpdate.Members != nil {
		var requested []string
		var result []string

		if len(userID) > 0 {
			requested = append(requested, userID[0])
		} else {
			requested = append(requested, AdminUserID)
		}

		for _, member := range *teamUpdate.Members {
			if member.ID == AdminUserID {
				continue
			}

			requested = append(requested, member.ID)
		}

		for _, member := range teamRead.Members {
			result = append(result, member.ID)
		}

		EqualOrEmpty(t, requested, result, "invalid team members")
	}

	return teamRead
}

func DeleteTeam(t *testing.T, id string, userID ...string) {
	resp := DoDeleteRequest(t, "/teams/"+id, userID...)
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		t.Errorf("team was not deleted")
	}
}

func JoinTeam(t *testing.T, id string, invitationCode string, userID ...string) {
	teamJoin := body.TeamJoin{
		InvitationCode: invitationCode,
	}

	resp := DoPostRequest(t, "/teams/"+id, teamJoin, userID...)
	_ = Parse[struct{}](t, resp)
}

func WithTeam(t *testing.T, teamCreate body.TeamCreate, userID ...string) body.TeamRead {
	var requestedMembers []string
	if len(userID) > 0 {
		requestedMembers = append(requestedMembers, userID[0])
	} else {
		requestedMembers = append(requestedMembers, AdminUserID)
	}

	for _, member := range teamCreate.Members {
		if member.ID == AdminUserID {
			continue
		}

		requestedMembers = append(requestedMembers, member.ID)
	}
	requestedResources := teamCreate.Resources

	var createdMembers []string
	var createdResources []string

	resp := DoPostRequest(t, "/teams", teamCreate, userID...)
	teamRead := Parse[body.TeamRead](t, resp)

	assert.Equal(t, teamCreate.Name, teamRead.Name, "invalid team name")
	assert.Equal(t, teamCreate.Description, teamRead.Description, "invalid team description")

	for _, member := range teamRead.Members {
		createdMembers = append(createdMembers, member.ID)
	}

	for _, resource := range teamRead.Resources {
		createdResources = append(createdResources, resource.ID)
	}

	EqualOrEmpty(t, requestedMembers, createdMembers, "invalid team members")
	EqualOrEmpty(t, requestedResources, createdResources, "invalid team resources")

	t.Cleanup(func() {
		resp = DoDeleteRequest(t, "/teams/"+teamRead.ID)
		if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
			assert.Fail(t, "team was not deleted")
		}
	})

	return teamRead
}

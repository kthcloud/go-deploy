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
		EqualOrEmpty(t, *teamUpdate.Resources, teamRead.Resources, "invalid team resources")
	}

	if teamUpdate.Members != nil {
		EqualOrEmpty(t, *teamUpdate.Members, teamRead.Members, "invalid team members")
	}

	return teamRead
}

func DeleteTeam(t *testing.T, id string, userID ...string) {
	resp := DoDeleteRequest(t, "/teams/"+id, userID...)
	if resp.StatusCode != 201 && resp.StatusCode != 404 {
		t.Errorf("team was not deleted")
	}
}

func JoinTeam(t *testing.T, id string, invitationCode string, userID ...string) {
	teamJoin := body.TeamJoin{
		InvitationCode: invitationCode,
	}

	resp := DoPostRequest(t, "/teams/"+id, teamJoin, userID...)
	assert.Equal(t, 200, resp.StatusCode, "team was not joined")
}

func WithTeam(t *testing.T, teamCreate body.TeamCreate) body.TeamRead {
	resp := DoPostRequest(t, "/teams", teamCreate)
	assert.Equal(t, 200, resp.StatusCode)

	var teamRead body.TeamRead
	err := ReadResponseBody(t, resp, &teamRead)
	assert.NoError(t, err, "team was not created")
	assert.Equal(t, teamCreate.Name, teamRead.Name, "invalid team name")
	assert.Equal(t, teamCreate.Description, teamRead.Description, "invalid team description")
	EqualOrEmpty(t, teamCreate.Resources, teamRead.Resources, "invalid team resources")
	EqualOrEmpty(t, teamCreate.Members, teamRead.Members, "invalid team members")

	t.Cleanup(func() {
		resp = DoDeleteRequest(t, "/teams/"+teamRead.ID)
		if resp.StatusCode != 201 && resp.StatusCode != 404 {
			t.Errorf("team was not deleted")
		}
	})

	return teamRead
}

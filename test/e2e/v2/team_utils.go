package v2

import (
	"net/http"
	"testing"

	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/test"
	"github.com/kthcloud/go-deploy/test/e2e"
	"github.com/stretchr/testify/assert"
)

const (
	TeamPath  = "/v2/teams/"
	TeamsPath = "/v2/teams"
)

func GetTeam(t *testing.T, id string, user ...string) body.TeamRead {
	resp := e2e.DoGetRequest(t, TeamPath+id, user...)
	return e2e.MustParse[body.TeamRead](t, resp)
}

func ListTeams(t *testing.T, query string, user ...string) []body.TeamRead {
	resp := e2e.DoGetRequest(t, TeamsPath+query, user...)
	return e2e.MustParse[[]body.TeamRead](t, resp)
}

func UpdateTeam(t *testing.T, id string, teamUpdate body.TeamUpdate, user ...string) body.TeamRead {
	resp := e2e.DoPostRequest(t, TeamPath+id, teamUpdate, user...)
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

		requested = append(requested, *teamUpdate.Resources...)

		for _, resource := range teamRead.Resources {
			result = append(result, resource.ID)
		}

		test.EqualOrEmpty(t, requested, result, "invalid team resources")
	}

	if teamUpdate.Members != nil {
		var requested []string
		var result []string

		if len(user) > 0 {
			requested = append(requested, e2e.GetUserID(user[0]))
		} else {
			requested = append(requested, model.TestAdminUserID)
		}

		for _, member := range *teamUpdate.Members {
			if member.ID == model.TestAdminUserID {
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

func DeleteTeam(t *testing.T, id string, user ...string) {
	resp := e2e.DoDeleteRequest(t, TeamPath+id, user...)
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		t.Errorf("team was not deleted")
	}
}

func JoinTeam(t *testing.T, id string, invitationCode string, user ...string) {
	teamJoin := body.TeamJoin{
		InvitationCode: invitationCode,
	}

	resp := e2e.DoPostRequest(t, TeamPath+id, teamJoin, user...)
	_ = e2e.MustParse[struct{}](t, resp)
}

func WithTeam(t *testing.T, teamCreate body.TeamCreate, user ...string) body.TeamRead {
	var requestedMembers []string
	if len(user) > 0 {
		requestedMembers = append(requestedMembers, e2e.GetUserID(user[0]))
	} else {
		requestedMembers = append(requestedMembers, model.TestPowerUserID)
	}

	for _, member := range teamCreate.Members {
		if member.ID == model.TestPowerUserID {
			continue
		}

		requestedMembers = append(requestedMembers, member.ID)
	}
	requestedResources := teamCreate.Resources

	var createdMembers []string
	var createdResources []string

	resp := e2e.DoPostRequest(t, TeamsPath, teamCreate, user...)
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
		e2e.DoDeleteRequest(t, TeamPath+teamRead.ID, user...)
	})

	return teamRead
}

package teams

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/dto/v2/body"
	"go-deploy/models/model"
	"go-deploy/test/e2e"
	"go-deploy/test/e2e/v2"
	"net/http"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	e2e.Setup()
	code := m.Run()
	e2e.Shutdown()
	os.Exit(code)
}

func TestCreateEmptyTeam(t *testing.T) {
	t.Parallel()

	requestBody := body.TeamCreate{
		Name:        e2e.GenName(),
		Description: e2e.GenName(),
		Resources:   nil,
		Members:     nil,
	}

	_ = v2.WithTeam(t, requestBody)
}

func TestCreateWithMembers(t *testing.T) {
	t.Parallel()

	requestBody := body.TeamCreate{
		Name:        e2e.GenName(),
		Description: e2e.GenName(),
		Resources:   nil,
		Members: []body.TeamMemberCreate{
			{ID: model.TestDefaultUserID, TeamRole: model.TeamMemberRoleAdmin},
		},
	}

	// Create team
	_ = v2.WithTeam(t, requestBody, e2e.PowerUser)

	// Fetch TestUser2's teams
	teams := v2.ListTeams(t, "?userId="+model.TestDefaultUserID, e2e.DefaultUser)
	found := false
	for _, team := range teams {
		if team.Name == requestBody.Name {
			found = true
			break
		}
	}
	assert.True(t, found, "user has no teams")
}

func TestCreateWithResources(t *testing.T) {
	t.Parallel()

	resource, _ := v2.WithDeployment(t, body.DeploymentCreate{
		Name: e2e.GenName(),
	})

	requestBody := body.TeamCreate{
		Name:        e2e.GenName(),
		Description: e2e.GenName(),
		Resources:   []string{resource.ID},
		Members:     nil,
	}

	// Create Team
	team := v2.WithTeam(t, requestBody)

	// Fetch deployment's teams
	resource = v2.GetDeployment(t, resource.ID)
	assert.EqualValues(t, []string{team.ID}, resource.Teams, "invalid teams on model")
}

func TestCreateFull(t *testing.T) {
	t.Parallel()

	resource, _ := v2.WithDeployment(t, body.DeploymentCreate{
		Name: e2e.GenName(),
	})

	requestBody := body.TeamCreate{
		Name:        e2e.GenName(),
		Description: e2e.GenName(),
		Resources:   []string{resource.ID},
		Members:     []body.TeamMemberCreate{{ID: model.TestDefaultUserID, TeamRole: model.TeamMemberRoleAdmin}},
	}

	// Create team
	team := v2.WithTeam(t, requestBody)

	// Fetch TestUser2's teams
	teams := v2.ListTeams(t, "?userId="+model.TestDefaultUserID, e2e.DefaultUser)
	assert.NotEmpty(t, teams, "user has no teams")

	// Fetch deployment's teams
	resource = v2.GetDeployment(t, resource.ID)
	assert.EqualValues(t, []string{team.ID}, resource.Teams, "invalid teams on model")
}

func TestCreateWithInvitation(t *testing.T) {
	t.Parallel()

	team := v2.WithTeam(t, body.TeamCreate{
		Name:        e2e.GenName(),
		Description: e2e.GenName(),
		Resources:   nil,
		Members:     []body.TeamMemberCreate{{ID: model.TestDefaultUserID, TeamRole: model.TeamMemberRoleAdmin}},
	}, e2e.PowerUser)

	assert.Equal(t, 2, len(team.Members), "invalid number of members")

	found := false
	for _, member := range team.Members {
		if member.ID == model.TestDefaultUserID {
			assert.Equal(t, model.TeamMemberRoleAdmin, member.TeamRole, "invalid member role")
			assert.Equal(t, model.TeamMemberStatusInvited, member.MemberStatus, "invalid member status")

			found = true
			break
		}
	}

	if !found {
		assert.Fail(t, "user was not invited")
	}

	notifications := v2.ListNotifications(t, "?userId="+model.TestDefaultUserID, e2e.DefaultUser)
	assert.NotEmpty(t, notifications, "user has no notifications")

	found = false
	for _, notification := range notifications {
		if notification.Type == model.NotificationTeamInvite {
			for key, val := range notification.Content {
				if key == "id" && val == team.ID {
					return
				}
			}
		}
	}

	assert.Fail(t, "user has no team invite notification")
}

func TestJoin(t *testing.T) {
	t.Parallel()

	team := v2.WithTeam(t, body.TeamCreate{
		Name:        e2e.GenName(),
		Description: e2e.GenName(),
		Resources:   nil,
		Members:     []body.TeamMemberCreate{{ID: model.TestDefaultUserID, TeamRole: model.TeamMemberRoleAdmin}},
	}, e2e.PowerUser)

	notifications := v2.ListNotifications(t, "?userId="+model.TestDefaultUserID, e2e.DefaultUser)

	for _, notification := range notifications {
		if notification.Type == model.NotificationTeamInvite {
			for key, val := range notification.Content {
				if key == "id" && val == team.ID {
					code := notification.Content["code"].(string)
					v2.JoinTeam(t, team.ID, code, e2e.DefaultUser)
					return
				}
			}
		}
	}

	assert.Fail(t, "user has no team invite notification")
}

func TestJoinWithBadCode(t *testing.T) {
	t.Parallel()

	team := v2.WithTeam(t, body.TeamCreate{
		Name:        e2e.GenName(),
		Description: e2e.GenName(),
		Resources:   nil,
		Members:     []body.TeamMemberCreate{{ID: model.TestDefaultUserID, TeamRole: model.TeamMemberRoleAdmin}},
	}, e2e.PowerUser)

	resp := e2e.DoPostRequest(t, v2.TeamPath+team.ID, body.TeamJoin{InvitationCode: "bad-code"}, e2e.DefaultUser)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "bad code was not detected")
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	team := v2.WithTeam(t, body.TeamCreate{
		Name:        e2e.GenName(),
		Description: e2e.GenName(),
		Resources:   nil,
		Members:     nil,
	})

	requestBody := body.TeamUpdate{
		Name:        e2e.StrPtr(e2e.GenName("new-team")),
		Description: e2e.StrPtr(e2e.GenName("new-description")),
		Resources:   nil,
		Members:     nil,
	}

	v2.UpdateTeam(t, team.ID, requestBody)
}

func TestUpdateResources(t *testing.T) {
	t.Parallel()

	team := v2.WithTeam(t, body.TeamCreate{
		Name:        e2e.GenName(),
		Description: e2e.GenName(),
		Resources:   nil,
		Members:     nil,
	})

	resource, _ := v2.WithDeployment(t, body.DeploymentCreate{
		Name: e2e.GenName("deployment"),
	})

	requestBody := body.TeamUpdate{
		Name:        nil,
		Description: nil,
		Resources:   &[]string{resource.ID},
		Members:     nil,
	}

	v2.UpdateTeam(t, team.ID, requestBody)

	// Fetch deployment's teams
	resource = v2.GetDeployment(t, resource.ID)
	assert.EqualValues(t, []string{team.ID}, resource.Teams, "invalid teams on model")
}

func TestUpdateMembers(t *testing.T) {
	t.Parallel()

	team := v2.WithTeam(t, body.TeamCreate{
		Name:        e2e.GenName(),
		Description: e2e.GenName(),
		Resources:   nil,
		Members:     nil,
	})

	requestBody := body.TeamUpdate{
		Name:        nil,
		Description: nil,
		Resources:   nil,
		Members:     &[]body.TeamMemberUpdate{{ID: model.TestDefaultUserID, TeamRole: model.TeamMemberRoleAdmin}},
	}

	v2.UpdateTeam(t, team.ID, requestBody, e2e.PowerUser)

	// Fetch PowerUser's teams, even though it is not accepted, it should be in the list
	teams := v2.ListTeams(t, "?userId="+model.TestDefaultUserID, e2e.DefaultUser)
	assert.NotEmpty(t, teams, "user has no teams")
}

func TestDelete(t *testing.T) {
	t.Parallel()

	team := v2.WithTeam(t, body.TeamCreate{
		Name:        e2e.GenName(),
		Description: e2e.GenName(),
		Resources:   nil,
		Members:     nil,
	})

	v2.DeleteTeam(t, team.ID)
}

func TestDeleteAsNonOwner(t *testing.T) {
	t.Parallel()

	team := v2.WithTeam(t, body.TeamCreate{
		Name:        e2e.GenName(),
		Description: e2e.GenName(),
		Resources:   nil,
		Members:     nil,
	})

	resp := e2e.DoDeleteRequest(t, v2.TeamPath+team.ID, e2e.DefaultUser)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "team was deleted by non-owner member")
}

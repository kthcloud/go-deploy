package teams

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/body"
	notificationModels "go-deploy/models/sys/notification"
	teamModels "go-deploy/models/sys/team"
	"go-deploy/test/e2e"
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
	requestBody := body.TeamCreate{
		Name:        e2e.GenName(),
		Description: e2e.GenName(),
		Resources:   nil,
		Members:     nil,
	}

	_ = e2e.WithTeam(t, requestBody)
}

func TestCreateWithMembers(t *testing.T) {
	requestBody := body.TeamCreate{
		Name:        e2e.GenName(),
		Description: e2e.GenName(),
		Resources:   nil,
		Members: []body.TeamMemberCreate{
			{ID: e2e.PowerUserID, TeamRole: teamModels.MemberRoleAdmin},
		},
	}

	// Create team
	_ = e2e.WithTeam(t, requestBody)

	// Fetch TestUser2's teams
	teams := e2e.ListTeams(t, "?userId="+e2e.PowerUserID)
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
	resource, _ := e2e.WithDeployment(t, body.DeploymentCreate{
		Name: e2e.GenName(),
	})

	requestBody := body.TeamCreate{
		Name:        e2e.GenName(),
		Description: e2e.GenName(),
		Resources:   []string{resource.ID},
		Members:     nil,
	}

	// Create Team
	team := e2e.WithTeam(t, requestBody)

	// Fetch deployment's teams
	resource = e2e.GetDeployment(t, resource.ID)
	assert.EqualValues(t, []string{team.ID}, resource.Teams, "invalid teams on resource")

}

func TestCreateFull(t *testing.T) {
	resource, _ := e2e.WithDeployment(t, body.DeploymentCreate{
		Name: e2e.GenName(),
	})

	requestBody := body.TeamCreate{
		Name:        e2e.GenName(),
		Description: e2e.GenName(),
		Resources:   []string{resource.ID},
		Members:     []body.TeamMemberCreate{{ID: e2e.PowerUserID, TeamRole: teamModels.MemberRoleAdmin}},
	}

	// Create team
	team := e2e.WithTeam(t, requestBody)

	// Fetch TestUser2's teams
	teams := e2e.ListTeams(t, "?userId="+e2e.PowerUserID)
	assert.NotEmpty(t, teams, "user has no teams")

	// Fetch deployment's teams
	resource = e2e.GetDeployment(t, resource.ID)
	assert.EqualValues(t, []string{team.ID}, resource.Teams, "invalid teams on resource")
}

func TestCreateWithInvitation(t *testing.T) {
	team := e2e.WithTeam(t, body.TeamCreate{
		Name:        e2e.GenName(),
		Description: e2e.GenName(),
		Resources:   nil,
		Members:     []body.TeamMemberCreate{{ID: e2e.PowerUserID, TeamRole: teamModels.MemberRoleAdmin}},
	}, e2e.DefaultUserID)

	assert.Equal(t, 2, len(team.Members), "invalid number of members")

	found := false
	for _, member := range team.Members {
		if member.ID == e2e.PowerUserID {
			assert.Equal(t, teamModels.MemberRoleAdmin, member.TeamRole, "invalid member role")
			assert.Equal(t, teamModels.MemberStatusInvited, member.MemberStatus, "invalid member status")

			found = true
			break
		}
	}

	if !found {
		assert.Fail(t, "user was not invited")
	}

	notifications := e2e.ListNotifications(t, "?userId="+e2e.PowerUserID)
	assert.NotEmpty(t, notifications, "user has no notifications")

	found = false
	for _, notification := range notifications {
		if notification.Type == notificationModels.TypeTeamInvite {
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
	team := e2e.WithTeam(t, body.TeamCreate{
		Name:        e2e.GenName(),
		Description: e2e.GenName(),
		Resources:   nil,
		Members:     []body.TeamMemberCreate{{ID: e2e.PowerUserID, TeamRole: teamModels.MemberRoleAdmin}},
	}, e2e.DefaultUserID)

	notifications := e2e.ListNotifications(t, "?userId="+e2e.PowerUserID)

	for _, notification := range notifications {
		if notification.Type == notificationModels.TypeTeamInvite {
			for key, val := range notification.Content {
				if key == "id" && val == team.ID {
					code := notification.Content["code"].(string)
					e2e.JoinTeam(t, team.ID, code, e2e.PowerUserID)
					return
				}
			}
		}
	}

	assert.Fail(t, "user has no team invite notification")
}

func TestJoinWithBadCode(t *testing.T) {
	team := e2e.WithTeam(t, body.TeamCreate{
		Name:        e2e.GenName(),
		Description: e2e.GenName(),
		Resources:   nil,
		Members:     []body.TeamMemberCreate{{ID: e2e.DefaultUserID, TeamRole: teamModels.MemberRoleAdmin}},
	})

	resp := e2e.DoPostRequest(t, "/teams/"+team.ID, body.TeamJoin{InvitationCode: "bad-code"}, e2e.DefaultUserID)
	assert.Equal(t, 400, resp.StatusCode, "bad code was not detected")
}

func TestUpdate(t *testing.T) {
	team := e2e.WithTeam(t, body.TeamCreate{
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

	e2e.UpdateTeam(t, team.ID, requestBody)
}

func TestUpdateResources(t *testing.T) {
	team := e2e.WithTeam(t, body.TeamCreate{
		Name:        e2e.GenName(),
		Description: e2e.GenName(),
		Resources:   nil,
		Members:     nil,
	})

	resource, _ := e2e.WithDeployment(t, body.DeploymentCreate{
		Name: e2e.GenName("deployment"),
	})

	requestBody := body.TeamUpdate{
		Name:        nil,
		Description: nil,
		Resources:   &[]string{resource.ID},
		Members:     nil,
	}

	e2e.UpdateTeam(t, team.ID, requestBody)

	// Fetch deployment's teams
	resource = e2e.GetDeployment(t, resource.ID)
	assert.EqualValues(t, []string{team.ID}, resource.Teams, "invalid teams on resource")
}

func TestUpdateMembers(t *testing.T) {
	team := e2e.WithTeam(t, body.TeamCreate{
		Name:        e2e.GenName(),
		Description: e2e.GenName(),
		Resources:   nil,
		Members:     nil,
	})

	requestBody := body.TeamUpdate{
		Name:        nil,
		Description: nil,
		Resources:   nil,
		Members:     &[]body.TeamMemberUpdate{{ID: e2e.PowerUserID, TeamRole: teamModels.MemberRoleAdmin}},
	}

	e2e.UpdateTeam(t, team.ID, requestBody)

	// Fetch TestUser2's teams
	teams := e2e.ListTeams(t, "?userId="+e2e.PowerUserID)
	assert.NotEmpty(t, teams, "user has no teams")
}

func TestDelete(t *testing.T) {
	team := e2e.WithTeam(t, body.TeamCreate{
		Name:        e2e.GenName(),
		Description: e2e.GenName(),
		Resources:   nil,
		Members:     nil,
	})

	e2e.DeleteTeam(t, team.ID)
}

func TestDeleteAsNonOwner(t *testing.T) {
	team := e2e.WithTeam(t, body.TeamCreate{
		Name:        e2e.GenName(),
		Description: e2e.GenName(),
		Resources:   nil,
		Members:     nil,
	})

	resp := e2e.DoDeleteRequest(t, "/teams/"+team.ID, e2e.DefaultUserID)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "team was deleted by non-owner member")
}

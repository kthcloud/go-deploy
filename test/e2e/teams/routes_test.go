package teams

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/body"
	notification2 "go-deploy/models/sys/notification"
	teamModels "go-deploy/models/sys/team"
	"go-deploy/test/e2e"
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
		Name:        e2e.GenName("team"),
		Description: e2e.GenName("description"),
		Resources:   nil,
		Members:     nil,
	}

	_ = e2e.WithTeam(t, requestBody)
}

func TestCreateWithMembers(t *testing.T) {
	requestBody := body.TeamCreate{
		Name:        e2e.GenName("team"),
		Description: e2e.GenName("description"),
		Resources:   nil,
		Members: []body.TeamMemberCreate{
			{ID: e2e.PowerUserID, TeamRole: teamModels.MemberRoleAdmin},
		},
	}

	// Fetch TestUser2's teams
	teams := e2e.ListTeams(t, "?userID="+e2e.PowerUserID)
	assert.NotEmpty(t, teams, "user has no teams")

	_ = e2e.WithTeam(t, requestBody)
}

func TestCreateWithResources(t *testing.T) {
	resource, _ := e2e.WithDeployment(t, body.DeploymentCreate{
		Name: e2e.GenName("deployment"),
	})

	requestBody := body.TeamCreate{
		Name:        e2e.GenName("team"),
		Description: e2e.GenName("description"),
		Resources:   []string{resource.ID},
		Members:     nil,
	}

	// Fetch deployment's teams
	resource = e2e.GetDeployment(t, resource.ID)
	assert.EqualValues(t, []string{resource.ID}, resource.Teams, "invalid teams on resource")

	_ = e2e.WithTeam(t, requestBody)
}

func TestCreateFull(t *testing.T) {
	resource, _ := e2e.WithDeployment(t, body.DeploymentCreate{
		Name: e2e.GenName("deployment"),
	})

	requestBody := body.TeamCreate{
		Name:        e2e.GenName("team"),
		Description: e2e.GenName("description"),
		Resources:   []string{resource.ID},
		Members:     []body.TeamMemberCreate{{ID: e2e.PowerUserID, TeamRole: teamModels.MemberRoleAdmin}},
	}

	// Fetch TestUser2's teams
	teams := e2e.ListTeams(t, "?userID="+e2e.PowerUserID)
	assert.NotEmpty(t, teams, "user has no teams")

	// Fetch deployment's teams
	resource = e2e.GetDeployment(t, resource.ID)
	assert.EqualValues(t, []string{resource.ID}, resource.Teams, "invalid teams on resource")

	e2e.WithTeam(t, requestBody)
}

func TestCreateWithInvitation(t *testing.T) {
	team := e2e.WithTeam(t, body.TeamCreate{
		Name:        e2e.GenName("team"),
		Description: e2e.GenName("description"),
		Resources:   nil,
		Members:     []body.TeamMemberCreate{{ID: e2e.DefaultUserID, TeamRole: teamModels.MemberRoleAdmin}},
	})

	assert.Equal(t, 1, len(team.Members), "invalid number of members")
	assert.Equal(t, e2e.DefaultUserID, team.Members[0].ID, "invalid member ID")
	assert.Equal(t, teamModels.MemberRoleAdmin, team.Members[0].TeamRole, "invalid member role")
	assert.Equal(t, teamModels.MemberStatusInvited, team.Members[0].MemberStatus, "invalid member status")

	notifications := e2e.ListNotifications(t, "?userID="+e2e.DefaultUserID)
	assert.NotEmpty(t, notifications, "user has no notifications")

	found := false
	for _, notification := range notifications {
		if notification.UserID == e2e.PowerUserID && notification.Type == notification2.TypeTeamInvite {
			found = true
			break
		}
	}

	assert.True(t, found, "user has no team invite notification")
}

func TestJoin(t *testing.T) {
	team := e2e.WithTeam(t, body.TeamCreate{
		Name:        e2e.GenName("team"),
		Description: e2e.GenName("description"),
		Resources:   nil,
		Members:     []body.TeamMemberCreate{{ID: e2e.DefaultUserID, TeamRole: teamModels.MemberRoleAdmin}},
	})

	notifications := e2e.ListNotifications(t, "?userID="+e2e.DefaultUserID)

	for _, notification := range notifications {
		if notification.UserID == e2e.PowerUserID && notification.Type == notification2.TypeTeamInvite {
			code := notification.Content["code"].(string)
			e2e.JoinTeam(t, team.ID, code, e2e.DefaultUserID)
			return
		}
	}

	assert.Fail(t, "user has no team invite notification")
}

func TestJoinWithBadCode(t *testing.T) {
	team := e2e.WithTeam(t, body.TeamCreate{
		Name:        e2e.GenName("team"),
		Description: e2e.GenName("description"),
		Resources:   nil,
		Members:     []body.TeamMemberCreate{{ID: e2e.DefaultUserID, TeamRole: teamModels.MemberRoleAdmin}},
	})

	resp := e2e.DoPostRequest(t, "/teams/"+team.ID, body.TeamJoin{InvitationCode: "bad-code"}, e2e.DefaultUserID)
	assert.Equal(t, 400, resp.StatusCode, "bad code was not detected")
}

func TestUpdate(t *testing.T) {
	team := e2e.WithTeam(t, body.TeamCreate{
		Name:        e2e.GenName("team"),
		Description: e2e.GenName("description"),
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
		Name:        e2e.GenName("team"),
		Description: e2e.GenName("description"),
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
	assert.EqualValues(t, []string{resource.ID}, resource.Teams, "invalid teams on resource")

}

func TestUpdateMembers(t *testing.T) {
	team := e2e.WithTeam(t, body.TeamCreate{
		Name:        e2e.GenName("team"),
		Description: e2e.GenName("description"),
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
	teams := e2e.ListTeams(t, "?userID="+e2e.PowerUserID)
	assert.NotEmpty(t, teams, "user has no teams")
}

func TestDelete(t *testing.T) {
	team := e2e.WithTeam(t, body.TeamCreate{
		Name:        e2e.GenName("team"),
		Description: e2e.GenName("description"),
		Resources:   nil,
		Members:     nil,
	})

	e2e.DeleteTeam(t, team.ID)
}

func TestDeleteAsNonOwner(t *testing.T) {
	team := e2e.WithTeam(t, body.TeamCreate{
		Name:        e2e.GenName("team"),
		Description: e2e.GenName("description"),
		Resources:   nil,
		Members: []body.TeamMemberCreate{
			{ID: e2e.PowerUserID, TeamRole: teamModels.MemberRoleAdmin},
		},
	})

	resp := e2e.DoDeleteRequest(t, "/teams/"+team.ID, e2e.DefaultUserID)
	assert.Equal(t, 403, resp.StatusCode, "team was deleted by non-owner member")
}

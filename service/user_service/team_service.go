package user_service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/models/dto/body"
	deploymentModels "go-deploy/models/sys/deployment"
	notificationModels "go-deploy/models/sys/notification"
	teamModels "go-deploy/models/sys/team"
	vmModels "go-deploy/models/sys/vm"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/notification_service"
	"go-deploy/utils"
	"time"
)

// GetTeam gets a team
//
// It uses service.AuthInfo to only return the resource the requesting user has access to
func (c *Client) GetTeam(id string, opts ...GetTeamOpts) (*teamModels.Team, error) {
	_ = service.GetFirstOrDefault(opts)

	tmc := teamModels.New()

	if c.Auth != nil && !c.Auth.IsAdmin {
		tmc.WithUserID(c.Auth.UserID)
	}

	return c.Team(id, tmc)
}

// ListTeams lists teams
//
// It uses service.AuthInfo to only return the resources the requesting user has access to
func (c *Client) ListTeams(opts ...ListTeamsOpts) ([]teamModels.Team, error) {
	o := service.GetFirstOrDefault(opts)

	tmc := teamModels.New()

	if o.Pagination != nil {
		tmc.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	var effectiveUserID string
	if o.UserID != "" {
		// Specific user's teams are requested
		if c.Auth == nil || c.Auth.UserID == o.UserID || c.Auth.IsAdmin {
			effectiveUserID = o.UserID
		} else {
			// User cannot access the other user's resources
			return nil, nil
		}
	} else {
		// All teams are requested
		if c.Auth != nil && !c.Auth.IsAdmin {
			effectiveUserID = c.Auth.UserID
		}
	}

	if effectiveUserID != "" {
		tmc.WithUserID(effectiveUserID)
	}

	if o.ResourceID != "" {
		tmc.WithResourceID(o.ResourceID)
	}

	return c.Teams(tmc)
}

// CreateTeam creates a new team
//
// Notifications are sent out if the owner of the team is not admin
func (c *Client) CreateTeam(id, ownerID string, dtoCreateTeam *body.TeamCreate) (*teamModels.Team, error) {
	params := &teamModels.CreateParams{}
	params.FromDTO(dtoCreateTeam, ownerID,
		func(resourceID string) *teamModels.Resource { return c.getResourceIfAccessible(resourceID) },
		func(memberDTO *body.TeamMemberCreate) *teamModels.Member {
			return c.createMemberIfAccessible(nil, memberDTO.ID)
		},
	)

	team, err := teamModels.New().Create(id, ownerID, params)
	if err != nil {
		if errors.Is(err, teamModels.NameTakenErr) {
			return nil, sErrors.TeamNameTakenErr
		}

		return nil, err
	}

	// Send invitations to every member that received an invitation code
	for _, member := range params.MemberMap {
		if member.InvitationCode != "" {
			err = createInvitationNotification(member.ID, team.ID, team.Name, member.InvitationCode)
			if err != nil {
				return nil, err
			}
		}
	}

	return team, nil
}

// UpdateTeam updates a team
//
// It uses service.AuthInfo to only update the resource the requesting user has access to
// Notifications are sent out if the owner of the team is not admin
func (c *Client) UpdateTeam(id string, dtoUpdateTeam *body.TeamUpdate) (*teamModels.Team, error) {
	team, err := teamModels.New().GetByID(id)
	if err != nil {
		return nil, err
	}

	if team == nil {
		return nil, nil
	}

	if c.Auth != nil && team.OwnerID != c.Auth.UserID && !c.Auth.IsAdmin && !team.HasMember(c.Auth.UserID) {
		return nil, nil
	}

	params := &teamModels.UpdateParams{}
	params.FromDTO(dtoUpdateTeam, team.GetMember(team.OwnerID),
		func(resourceID string) *teamModels.Resource { return c.getResourceIfAccessible(resourceID) },
		func(memberDTO *body.TeamMemberUpdate) *teamModels.Member {
			return c.createMemberIfAccessible(team, memberDTO.ID)
		},
	)

	// If new user, set timestamp
	if params.MemberMap != nil {
		for _, member := range *params.MemberMap {
			// Don't invite users that have already joined
			if existing := team.GetMember(member.ID); existing != nil && existing.MemberStatus == teamModels.MemberStatusJoined {
				continue
			}

			// Send notification to new users
			if member.InvitationCode != "" {
				err = createInvitationNotification(member.ID, team.ID, team.Name, member.InvitationCode)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// If new resource, set timestamp
	if params.ResourceMap != nil {
		for _, resource := range *params.ResourceMap {
			if existing := team.GetResource(resource.ID); existing != nil {
				resource.AddedAt = existing.AddedAt
			} else {
				resource.AddedAt = time.Now()
			}

			(*params.ResourceMap)[resource.ID] = resource
		}
	}

	tmc := teamModels.New()

	err = tmc.UpdateWithParams(id, params)
	if err != nil {
		return nil, err
	}

	return c.RefreshTeam(id, tmc)
}

// DeleteTeam deletes a team
//
// It uses service.AuthInfo to only delete the resource the requesting user has access to
func (c *Client) DeleteTeam(id string) error {
	tmc := teamModels.New()

	if c.Auth != nil && !c.Auth.IsAdmin {
		tmc.WithOwnerID(c.Auth.UserID)
	}

	exists, err := tmc.ExistsByID(id)
	if err != nil {
		return err
	}

	if !exists {
		return sErrors.TeamNotFoundErr
	}

	err = notificationModels.New().FilterContent("id", id).Delete()
	if err != nil {
		return err
	}

	return tmc.DeleteByID(id)
}

// JoinTeam joins a team
//
// It uses service.AuthInfo to get the user ID
func (c *Client) JoinTeam(id string, dtoTeamJoin *body.TeamJoin) (*teamModels.Team, error) {
	if c.Auth == nil {
		return nil, nil
	}

	params := &teamModels.JoinParams{}
	params.FromDTO(dtoTeamJoin)

	tmc := teamModels.New()
	team, err := tmc.GetByID(id)
	if err != nil {
		return nil, err
	}

	if team == nil {
		return nil, nil
	}

	if team.GetMemberMap()[c.Auth.UserID].MemberStatus != teamModels.MemberStatusInvited {
		return team, sErrors.NotInvitedErr
	}

	if team.GetMemberMap()[c.Auth.UserID].InvitationCode != params.InvitationCode {
		return nil, sErrors.BadInviteCodeErr
	}

	updatedMember := team.GetMemberMap()[c.Auth.UserID]
	updatedMember.MemberStatus = teamModels.MemberStatusJoined
	updatedMember.JoinedAt = time.Now()

	err = tmc.UpdateMember(id, c.Auth.UserID, &updatedMember)
	if err != nil {
		return nil, err
	}

	nmc := notificationModels.New().WithUserID(c.Auth.UserID).FilterContent("id", id).WithType(notificationModels.TypeTeamInvite)
	err = nmc.MarkReadAndCompleted()
	if err != nil {
		return nil, err
	}

	return c.RefreshTeam(id, tmc)
}

func (c *Client) getResourceIfAccessible(resourceID string) *teamModels.Resource {
	// Try to fetch deployment
	dClient := deploymentModels.New()
	vClient := vmModels.New()

	if c.Auth != nil && !c.Auth.IsAdmin {
		dClient.RestrictToOwner(c.Auth.UserID)
		vClient.RestrictToOwner(c.Auth.UserID)
	}

	isOwner, err := dClient.ExistsByID(resourceID)
	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("failed to fetch deployment when checking user access when creating team: %w", err))
		return nil
	}

	if isOwner {
		return &teamModels.Resource{
			ID:      resourceID,
			Type:    teamModels.ResourceTypeDeployment,
			AddedAt: time.Now(),
		}
	}

	// Try to fetch vm
	isOwner, err = vmModels.New().ExistsByID(resourceID)
	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("failed to fetch vm when checking user access when creating team: %w", err))
		return nil
	}

	if isOwner {
		return &teamModels.Resource{
			ID:      resourceID,
			Type:    teamModels.ResourceTypeVM,
			AddedAt: time.Now(),
		}
	}

	return nil
}

func (c *Client) createMemberIfAccessible(current *teamModels.Team, memberID string) *teamModels.Member {
	if current != nil {
		if existing := current.GetMember(memberID); existing != nil {
			existing.TeamRole = teamModels.MemberRoleAdmin
			return existing
		}
	}

	member := &teamModels.Member{
		ID:       memberID,
		TeamRole: teamModels.MemberRoleAdmin,
		AddedAt:  time.Now(),
	}

	if c.Auth == nil || c.Auth.IsAdmin {
		member.MemberStatus = teamModels.MemberStatusJoined
		member.JoinedAt = time.Now()
	} else {
		member.MemberStatus = teamModels.MemberStatusInvited
		member.InvitationCode = createInvitationCode()
	}

	return member
}

func createInvitationNotification(userID, teamID, teamName, invitationCode string) error {
	_, err := notification_service.New().Create(uuid.NewString(), userID, &notificationModels.CreateParams{
		Type: notificationModels.TypeTeamInvite,
		Content: map[string]interface{}{
			"id":   teamID,
			"name": teamName,
			"code": invitationCode,
		},
	})

	return err
}

func createInvitationCode() string {
	return utils.HashString(uuid.NewString())
}

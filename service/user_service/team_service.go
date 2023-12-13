package user_service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/models/dto/body"
	deploymentModel "go-deploy/models/sys/deployment"
	notificationModel "go-deploy/models/sys/notification"
	teamModels "go-deploy/models/sys/team"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/notification_service"
	"go-deploy/utils"
	"sort"
	"time"
)

func (c *Client) GetTeam(id string, opts *GetTeamOpts) (*teamModels.Team, error) {
	teamClient := teamModels.New()

	if c.Auth != nil && !c.Auth.IsAdmin {
		teamClient.WithUserID(c.Auth.UserID)
	}

	team, err := teamClient.GetByID(id)
	if err != nil {
		return nil, err
	}

	return team, nil
}

func (c *Client) ListTeams(opts *ListTeamsOpts) ([]teamModels.Team, error) {
	teamClient := teamModels.New()

	if opts.Pagination != nil {
		teamClient.WithPagination(opts.Pagination.Page, opts.Pagination.PageSize)
	}

	var effectiveUserID string
	if opts.UserID != "" {
		// Specific user's teams are requested
		if c.Auth == nil || c.Auth.UserID == opts.UserID || c.Auth.IsAdmin {
			effectiveUserID = opts.UserID
		} else {
			effectiveUserID = c.Auth.UserID
		}
	} else {
		// All teams are requested
		if c.Auth != nil && !c.Auth.IsAdmin {
			effectiveUserID = c.Auth.UserID
		}
	}

	if effectiveUserID != "" {
		teamClient.WithUserID(effectiveUserID)
	}

	teams, err := teamClient.List()
	if err != nil {
		return nil, err
	}

	sort.Slice(teams, func(i, j int) bool {
		return teams[i].CreatedAt.After(teams[j].CreatedAt)
	})

	return teams, nil
}

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

func (c *Client) UpdateTeam(id string, dtoUpdateTeam *body.TeamUpdate) (*teamModels.Team, error) {
	team, err := teamModels.New().GetByID(id)
	if err != nil {
		return nil, err
	}

	if team == nil {
		return nil, sErrors.TeamNotFoundErr
	}

	if c.Auth != nil && team.OwnerID != c.Auth.UserID && !c.Auth.IsAdmin && !team.HasMember(c.Auth.UserID) {
		return nil, sErrors.TeamNotFoundErr
	}

	params := &teamModels.UpdateParams{}
	params.FromDTO(dtoUpdateTeam, team.OwnerID,
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

	err = teamModels.New().UpdateWithParams(id, params)
	if err != nil {
		return nil, err
	}

	afterUpdate, err := teamModels.New().GetByID(id)
	if err != nil {
		return nil, err
	}

	if afterUpdate == nil {
		return nil, sErrors.TeamNotFoundErr
	}

	err = teamModels.New().MarkUpdated(id)
	if err != nil {
		return nil, err
	}

	return afterUpdate, nil
}

func (c *Client) DeleteTeam(id string) error {
	teamClient := teamModels.New()

	if c.Auth != nil && !c.Auth.IsAdmin {
		teamClient.WithOwnerID(c.Auth.UserID)
	}

	if exists, err := teamClient.ExistsByID(id); !exists || err != nil {
		return sErrors.TeamNotFoundErr
	}

	return teamClient.DeleteByID(id)
}

func (c *Client) JoinTeam(id string, dtoTeamJoin *body.TeamJoin, auth *service.AuthInfo) (*teamModels.Team, error) {
	params := &teamModels.JoinParams{}
	params.FromDTO(dtoTeamJoin)

	teamClient := teamModels.New()
	team, err := teamClient.GetByID(id)
	if err != nil {
		return nil, err
	}

	if team == nil {
		return nil, sErrors.TeamNotFoundErr
	}

	if team.GetMemberMap()[auth.UserID].MemberStatus != teamModels.MemberStatusInvited {
		return team, sErrors.NotInvitedErr
	}

	if team.GetMemberMap()[auth.UserID].InvitationCode != params.InvitationCode {
		return nil, sErrors.BadInviteCodeErr
	}

	updatedMember := team.GetMemberMap()[auth.UserID]
	updatedMember.MemberStatus = teamModels.MemberStatusJoined
	updatedMember.JoinedAt = time.Now()

	err = teamClient.UpdateMember(id, auth.UserID, &updatedMember)
	if err != nil {
		return nil, err
	}

	return teamClient.GetByID(id)
}

func (c *Client) getResourceIfAccessible(resourceID string) *teamModels.Resource {
	// try to fetch deployment
	dClient := deploymentModel.New()
	vClient := vmModel.New()

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

	// try to fetch vm
	isOwner, err = vmModel.New().ExistsByID(resourceID)
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
	return notification_service.CreateNotification(uuid.NewString(), userID, &notificationModel.CreateParams{
		Type: notificationModel.TypeTeamInvite,
		Content: map[string]interface{}{
			"id":   teamID,
			"name": teamName,
			"code": invitationCode,
		},
	})
}

func createInvitationCode() string {
	return utils.HashString(uuid.NewString())
}
